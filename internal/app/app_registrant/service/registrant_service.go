package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"rakit-tiket-be/internal/app/app_registrant/dao"
	"rakit-tiket-be/internal/pkg/payment"
	pubEntity "rakit-tiket-be/pkg/entity"
	eventEntity "rakit-tiket-be/pkg/entity/app_event"
	orderEntity "rakit-tiket-be/pkg/entity/app_order"
	regEntity "rakit-tiket-be/pkg/entity/app_registrant"
	ticketEntity "rakit-tiket-be/pkg/entity/app_ticket"
	model "rakit-tiket-be/pkg/model/app_registrant"
	"rakit-tiket-be/pkg/util"
)

type RegistrantService interface {
	Register(ctx context.Context, req model.RegisterRequest) (*model.RegisterResponse, error)
}

type registrantService struct {
	log            util.LogUtil
	sqlDB          *sql.DB
	paymentFactory *payment.PaymentFactory
}

func MakeRegistrantService(log util.LogUtil, sqlDB *sql.DB, paymentFactory *payment.PaymentFactory) RegistrantService {
	return registrantService{
		log:            log,
		sqlDB:          sqlDB,
		paymentFactory: paymentFactory,
	}
}

func (s registrantService) Register(ctx context.Context, req model.RegisterRequest) (*model.RegisterResponse, error) {
	totalRequestedTickets := 1 + len(req.Attendees)

	// Grouping tiket untuk efisiensi atomic update
	ticketQtyMap := make(map[string]int)
	ticketQtyMap[string(req.Registrant.TicketID)]++

	var ticketIDs []string
	ticketIDs = append(ticketIDs, string(req.Registrant.TicketID))

	for _, att := range req.Attendees {
		ticketQtyMap[string(att.TicketID)]++
		ticketIDs = append(ticketIDs, string(att.TicketID))
	}

	dbTrx := dao.NewTransactionRegistrant(ctx, s.log, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	// Cari Order berdasarkan OrderNumber
	tickets, err := dbTrx.GetTicketDAO().Search(ctx, ticketEntity.TicketQuery{IDs: ticketIDs})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tickets: %v", err)
	}

	// Buat Map Tiket dan Pastikan Semua Tiket dari Event yang Sama
	ticketMap := make(map[string]ticketEntity.Ticket)
	var eventID pubEntity.UUID

	for i, t := range tickets {
		ticketMap[string(t.ID)] = t
		if i == 0 {
			eventID = t.EventID
		} else if t.EventID != eventID {
			return nil, errors.New("semua tiket dalam satu transaksi harus berasal dari event yang sama")
		}
	}

	// Ambil Konfigurasi Event
	events, err := dbTrx.GetEventDAO().Search(ctx, eventEntity.EventQuery{IDs: []string{string(eventID)}})
	if err != nil || len(events) == 0 {
		return nil, errors.New("event tidak ditemukan")
	}
	eventData := events[0]

	// Validasi Dinamis Maksimal Tiket Berdasarkan Event
	if totalRequestedTickets > eventData.MaxTicketPerTx {
		return nil, fmt.Errorf("maksimal %d tiket per registrasi untuk event ini", eventData.MaxTicketPerTx)
	}

	// ATOMIC BOOKING STOCK & Mapping Payment
	var totalCost float64
	var paymentItems []payment.Item

	for tID, qty := range ticketQtyMap {
		ticketData, exists := ticketMap[tID]
		if !exists {
			return nil, fmt.Errorf("tiket %s tidak ditemukan", tID)
		}

		// A. Eksekusi Atomic Booking!
		err := dbTrx.GetTicketDAO().BookStock(ctx, pubEntity.UUID(tID), qty)
		if err != nil {
			return nil, fmt.Errorf("stok tiket %s tidak mencukupi (habis)", ticketData.Title)
		}

		// B. Hitung Harga
		totalCost += ticketData.Price * float64(qty)
		paymentItems = append(paymentItems, payment.Item{
			ID:       string(ticketData.ID),
			Name:     ticketData.Title,
			Price:    ticketData.Price,
			Quantity: qty,
		})
	}

	// Generate Identifier Dinamis (Menggunakan Prefix dari Event)
	now := time.Now()
	registrantID := pubEntity.MakeUUID(req.Registrant.Email, now.String())
	orderID := pubEntity.MakeUUID("ORDER", req.Registrant.Email, now.String())

	// Default fallback misal prefix kosong, namun di database sudah 'NOT NULL'
	prefix := eventData.TicketPrefixCode
	if prefix == "" {
		prefix = "TKT"
	}

	uniqueCode := fmt.Sprintf("%s-%d-%05d", prefix, now.Year(), now.Unix()%100000)
	orderNumber := fmt.Sprintf("%s%d-%05d", prefix, now.Year(), now.Unix()%100000)

	var regBirthdate *time.Time
	if req.Registrant.Birthdate != nil && *req.Registrant.Birthdate != "" {
		t, _ := time.Parse("2006-01-02", *req.Registrant.Birthdate)
		regBirthdate = &t
	}

	// Insert Data Registrant
	registrant := regEntity.Registrant{
		ID:           registrantID,
		EventID:      eventID,
		UniqueCode:   uniqueCode,
		TicketID:     &req.Registrant.TicketID,
		Name:         req.Registrant.Name,
		Email:        req.Registrant.Email,
		Phone:        req.Registrant.Phone,
		Gender:       req.Registrant.Gender,
		Birthdate:    regBirthdate,
		TotalCost:    totalCost,
		TotalTickets: totalRequestedTickets,
		Status:       "pending",
	}
	registrant.CreatedAt = now

	if err := dbTrx.GetRegistrantDAO().Insert(ctx, []regEntity.Registrant{registrant}); err != nil {
		return nil, err
	}

	// Insert Data Attendees
	var attendees []regEntity.Attendee
	for _, att := range req.Attendees {
		var attBirthdate *time.Time
		if att.Birthdate != nil && *att.Birthdate != "" {
			t, _ := time.Parse("2006-01-02", *att.Birthdate)
			attBirthdate = &t
		}

		attendees = append(attendees, regEntity.Attendee{
			ID:           pubEntity.MakeUUID(att.Name, string(att.TicketID), now.String()),
			EventID:      eventID,
			RegistrantID: registrantID,
			TicketID:     att.TicketID,
			Name:         att.Name,
			Gender:       att.Gender,
			Birthdate:    attBirthdate,
		})
	}

	if len(attendees) > 0 {
		if err := dbTrx.GetAttendeeDAO().Insert(ctx, attendees); err != nil {
			return nil, err
		}
	}

	// Panggil Payment Gateway via Factory
	paymentProvider, err := s.paymentFactory.GetProvider(payment.GatewayMidtrans)
	if err != nil {
		return nil, err
	}

	paymentReq := payment.CreateTransactionRequest{
		OrderID:       orderNumber,
		Amount:        totalCost,
		Customer:      payment.Customer{Name: registrant.Name, Email: registrant.Email, Phone: registrant.Phone},
		Items:         paymentItems,
		ExpiryMinutes: 30,
	}

	paymentResp, err := paymentProvider.CreateTransaction(ctx, paymentReq)
	if err != nil {
		return nil, fmt.Errorf("payment gateway error: %v", err)
	}

	// Insert Data Order dengan URL Midtrans
	gateway := string(payment.GatewayMidtrans)
	expiresAt := now.Add(30 * time.Minute)

	order := orderEntity.Order{
		ID:             orderID,
		EventID:        eventID,
		RegistrantID:   registrantID,
		OrderNumber:    orderNumber,
		Amount:         totalCost,
		Currency:       "IDR",
		PaymentGateway: &gateway,
		PaymentStatus:  "pending",
		PaymentToken:   &paymentResp.Token,
		PaymentURL:     &paymentResp.RedirectURL,
		ExpiresAt:      &expiresAt,
	}
	order.CreatedAt = now

	if err := dbTrx.GetOrderDAO().Insert(ctx, []orderEntity.Order{order}); err != nil {
		return nil, err
	}

	if err := dbTrx.GetSqlTx().Commit(); err != nil {
		return nil, err
	}

	return &model.RegisterResponse{
		Order: model.OrderInfo{
			OrderID:       string(order.ID),
			OrderNumber:   order.OrderNumber,
			Amount:        order.Amount,
			Currency:      order.Currency,
			PaymentStatus: order.PaymentStatus,
			PaymentToken:  paymentResp.Token,
			RedirectURL:   paymentResp.RedirectURL,
		},
		Registrant: model.RegistrantInfo{
			ID:         string(registrant.ID),
			UniqueCode: registrant.UniqueCode,
		},
	}, nil
}
