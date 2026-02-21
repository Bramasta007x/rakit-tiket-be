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
	orderEntity "rakit-tiket-be/pkg/entity/app_order"
	regEntity "rakit-tiket-be/pkg/entity/app_registrant"
	ticketEntity "rakit-tiket-be/pkg/entity/app_ticket"
	model "rakit-tiket-be/pkg/model/app_registrant"
)

type RegistrantService interface {
	Register(ctx context.Context, req model.RegisterRequest) (*model.RegisterResponse, error)
}

type registrantService struct {
	sqlDB          *sql.DB
	paymentFactory *payment.PaymentFactory
}

func MakeRegistrantService(sqlDB *sql.DB, paymentFactory *payment.PaymentFactory) RegistrantService {
	return registrantService{
		sqlDB:          sqlDB,
		paymentFactory: paymentFactory,
	}
}

func (s registrantService) Register(ctx context.Context, req model.RegisterRequest) (*model.RegisterResponse, error) {
	// 1. Validasi Maksimal Tiket
	totalRequestedTickets := 1 + len(req.Attendees)
	if totalRequestedTickets > 4 {
		return nil, errors.New("maksimal 4 tiket per registrasi (termasuk registrant)")
	}

	// 2. Hitung Kebutuhan Kuantitas Per Jenis Tiket (Grouping)
	ticketQtyMap := make(map[string]int)
	ticketQtyMap[string(req.Registrant.TicketID)]++
	var ticketIDs []string
	ticketIDs = append(ticketIDs, string(req.Registrant.TicketID))

	for _, att := range req.Attendees {
		ticketQtyMap[string(att.TicketID)]++
		ticketIDs = append(ticketIDs, string(att.TicketID))
	}

	// 3. Memulai Database Transaction sesuai Pattern Arsitektur
	dbTrx := dao.NewTransaction(ctx, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback() // Akan di-rollback jika fungsi berakhir sebelum Commit()

	// 4. Fetch Master Data Tickets
	tickets, err := dbTrx.GetTicketDAO().Search(ctx, ticketEntity.TicketQuery{IDs: ticketIDs})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tickets: %v", err)
	}

	ticketMap := make(map[string]ticketEntity.Ticket)
	for _, t := range tickets {
		ticketMap[string(t.ID)] = t
	}

	// 5. Atomic Stock Validation & Deduction (Mencegah Race Condition/Overselling)
	var totalCost float64
	var paymentItems []payment.Item

	for ticketIDStr, requestedQty := range ticketQtyMap {
		ticketData, exists := ticketMap[ticketIDStr]
		if !exists {
			return nil, fmt.Errorf("tiket dengan ID %s tidak valid", ticketIDStr)
		}

		// Eksekusi potong stok langsung di DB secara atomic
		err := dbTrx.GetTicketDAO().DecreaseRemaining(ctx, pubEntity.UUID(ticketIDStr), requestedQty)
		if err != nil {
			return nil, fmt.Errorf("gagal memproses tiket %s: %v", ticketData.Title, err)
		}

		// Kalkulasi Harga & Mapping untuk Midtrans
		totalCost += ticketData.Price * float64(requestedQty)
		paymentItems = append(paymentItems, payment.Item{
			ID:       string(ticketData.ID),
			Name:     ticketData.Title,
			Price:    ticketData.Price,
			Quantity: requestedQty,
		})
	}

	// 6. Generate Relation IDs & Unique Codes di Service Layer
	now := time.Now()
	// Gunakan pubEntity.MakeUUID atau util generator sesuai preferensimu
	registrantID := pubEntity.MakeUUID(req.Registrant.Email, req.Registrant.Name, now.String())
	orderID := pubEntity.MakeUUID("ORDER", req.Registrant.Email, now.String())

	// Format: JMF-2026-00123
	uniqueCode := fmt.Sprintf("JMF-%d-%s", now.Year(), "TEST")
	orderNumber := fmt.Sprintf("JMF%d-%s", now.Year(), "TEST")

	var regBirthdate *time.Time
	if req.Registrant.Birthdate != nil && *req.Registrant.Birthdate != "" {
		t, _ := time.Parse("2006-01-02", *req.Registrant.Birthdate)
		regBirthdate = &t
	}

	// 7. Siapkan dan Insert Data Registrant
	registrant := regEntity.Registrant{
		ID:           registrantID,
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
		return nil, fmt.Errorf("failed to insert registrant: %v", err)
	}

	// 8. Siapkan dan Insert Data Attendees
	var attendees []regEntity.Attendee
	for _, att := range req.Attendees {
		var attBirthdate *time.Time
		if att.Birthdate != nil && *att.Birthdate != "" {
			t, _ := time.Parse("2006-01-02", *att.Birthdate)
			attBirthdate = &t
		}

		attendees = append(attendees, regEntity.Attendee{
			ID:           pubEntity.MakeUUID(att.Name, string(att.TicketID), now.String()),
			RegistrantID: registrantID, // Map ke ID Registrant yang sudah di-generate
			TicketID:     att.TicketID,
			Name:         att.Name,
			Gender:       att.Gender,
			Birthdate:    attBirthdate,
		})
	}

	if len(attendees) > 0 {
		if err := dbTrx.GetAttendeeDAO().Insert(ctx, attendees); err != nil {
			return nil, fmt.Errorf("failed to insert attendees: %v", err)
		}
	}

	// 9. Call Payment Gateway via Strategy Factory
	paymentProvider, err := s.paymentFactory.GetProvider(payment.GatewayMidtrans)
	if err != nil {
		return nil, fmt.Errorf("payment gateway error: %v", err)
	}

	paymentReq := payment.CreateTransactionRequest{
		OrderID: orderNumber,
		Amount:  totalCost,
		Customer: payment.Customer{
			Name:  registrant.Name,
			Email: registrant.Email,
			Phone: registrant.Phone,
		},
		Items: paymentItems,
	}

	paymentResp, err := paymentProvider.CreateTransaction(ctx, paymentReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment transaction: %v", err)
	}

	// 10. Siapkan dan Insert Data Order dengan Token Payment
	gateway := string(payment.GatewayMidtrans)
	order := orderEntity.Order{
		ID:             orderID,
		RegistrantID:   registrantID, // Map ke ID Registrant
		OrderNumber:    orderNumber,
		Amount:         totalCost,
		Currency:       "IDR",
		PaymentGateway: &gateway,
		PaymentStatus:  "pending",
		PaymentToken:   &paymentResp.Token,
		PaymentURL:     &paymentResp.RedirectURL,
	}
	order.CreatedAt = now

	if err := dbTrx.GetOrderDAO().Insert(ctx, []orderEntity.Order{order}); err != nil {
		return nil, fmt.Errorf("failed to insert order: %v", err)
	}

	// 11. Final Commit Transaction
	if err := dbTrx.GetSqlTx().Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	// 12. Return Final Response
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
