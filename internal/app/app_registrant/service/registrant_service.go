package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"rakit-tiket-be/internal/app/app_registrant/dao"
	ticketDao "rakit-tiket-be/internal/app/app_ticket/dao"
	"rakit-tiket-be/internal/pkg/payment"

	orderEntity "rakit-tiket-be/pkg/entity/app_order"
	regEntity "rakit-tiket-be/pkg/entity/app_registrant"
	ticketEntity "rakit-tiket-be/pkg/entity/app_ticket"
	"rakit-tiket-be/pkg/model/app_registrant"
	"rakit-tiket-be/pkg/util"

	"go.uber.org/zap"
)

type RegistrantService interface {
	Register(ctx context.Context, req app_registrant.RegisterRequest) (*app_registrant.RegisterResponse, error)
}

type registrantService struct {
	log            util.LogUtil
	sqlDB          *sql.DB
	ticketDao      ticketDao.TicketDAO
	paymentFactory *payment.PaymentFactory
}

func MakeRegistrantService(log util.LogUtil, sqlDB *sql.DB, ticketDao ticketDao.TicketDAO, paymentFactory *payment.PaymentFactory) RegistrantService {
	return &registrantService{
		log:            log,
		sqlDB:          sqlDB,
		ticketDao:      ticketDao,
		paymentFactory: paymentFactory,
	}
}

func (s *registrantService) Register(ctx context.Context, req app_registrant.RegisterRequest) (*app_registrant.RegisterResponse, error) {
	dbTrx := dao.NewTransaction(ctx, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	// 1. Validasi Maksimal 4 Tiket
	if 1+len(req.Attendees) > 4 {
		return nil, errors.New("maksimal 4 tiket per registrasi (termasuk registrant)")
	}

	// 2. Kumpulkan semua Ticket IDs yang dibeli untuk validasi DB
	var ticketIDs []string
	ticketIDs = append(ticketIDs, string(req.Registrant.TicketID))
	for _, att := range req.Attendees {
		ticketIDs = append(ticketIDs, string(att.TicketID))
	}

	// 3. Fetch Tickets dari DB
	tickets, err := s.ticketDao.Search(ctx, ticketEntity.TicketQuery{IDs: ticketIDs})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tickets: %v", err)
	}

	// Buat Map untuk mempercepat pencarian tiket
	ticketMap := make(map[string]ticketEntity.Ticket)
	for _, t := range tickets {
		ticketMap[string(t.ID)] = t
	}

	// 4. Hitung Total Cost & Validasi Stok Tersisa
	var totalCost float64
	var ticketsInvolved []ticketEntity.Ticket

	// Cek tiket registrant
	regTicket, exists := ticketMap[string(req.Registrant.TicketID)]
	if !exists {
		return nil, fmt.Errorf("tiket dengan ID %s tidak ditemukan", req.Registrant.TicketID)
	}
	if regTicket.Remaining <= 0 {
		return nil, fmt.Errorf("tiket %s sudah habis", regTicket.Title)
	}
	totalCost += regTicket.Price
	ticketsInvolved = append(ticketsInvolved, regTicket)

	// Cek tiket attendees
	for _, att := range req.Attendees {
		t, exists := ticketMap[string(att.TicketID)]
		if !exists {
			return nil, fmt.Errorf("tiket dengan ID %s tidak ditemukan", att.TicketID)
		}
		if t.Remaining <= 0 {
			return nil, fmt.Errorf("tiket %s sudah habis", t.Title)
		}
		totalCost += t.Price
		ticketsInvolved = append(ticketsInvolved, t)
	}

	totalTickets := len(ticketsInvolved)

	// 5. Generate Codes
	year := time.Now().Year()
	rand.Seed(time.Now().UnixNano())
	randomNum := rand.Intn(99999) + 1

	// Contoh: JMF-2026-00123
	uniqueCode := fmt.Sprintf("JMF-%d-%05d", year, randomNum)
	// Contoh: JMF2026-00123
	orderNumber := fmt.Sprintf("JMF%d-%05d", year, randomNum)

	// Parse Birthdate
	var regBirthdate *time.Time
	if req.Registrant.Birthdate != nil && *req.Registrant.Birthdate != "" {
		t, _ := time.Parse("2006-01-02", *req.Registrant.Birthdate)
		regBirthdate = &t
	}

	// 6. Insert Registrant
	registrant := regEntity.Registrant{
		UniqueCode:   uniqueCode,
		TicketID:     &req.Registrant.TicketID,
		Name:         req.Registrant.Name,
		Email:        req.Registrant.Email,
		Phone:        req.Registrant.Phone,
		Gender:       req.Registrant.Gender,
		Birthdate:    regBirthdate,
		TotalCost:    totalCost,
		TotalTickets: totalTickets,
		Status:       "pending",
	}
	err = dbTrx.GetRegistrantDAO().Insert(ctx, []regEntity.Registrant{registrant})
	if err != nil {
		return nil, err
	}
	// Ambil ID Registrant yang baru digenerate (di DAO Insert kita men-set ID-nya)
	// Idealnya method Insert DAO mengembalikan array registrant yang sudah ada ID-nya,
	// Jika tidak, kamu harus fetch lagi atau pastikan logic generate ID di DAO menempel pada object reference.
	// Asumsi logic DAO-mu mengupdate struct asli by pointer/index array.

	// 7. Insert Attendees
	var attendees []regEntity.Attendee
	for _, att := range req.Attendees {
		var attBirthdate *time.Time
		if att.Birthdate != nil && *att.Birthdate != "" {
			t, _ := time.Parse("2006-01-02", *att.Birthdate)
			attBirthdate = &t
		}

		attendees = append(attendees, regEntity.Attendee{
			RegistrantID: registrant.ID, // Hubungkan ke pembeli
			TicketID:     att.TicketID,
			Name:         att.Name,
			Gender:       att.Gender,
			Birthdate:    attBirthdate,
		})
	}

	if len(attendees) > 0 {
		err = dbTrx.GetAttendeeDAO().Insert(ctx, attendees)
		if err != nil {
			return nil, err
		}
	}

	// 8. Insert Order
	gateway := string(payment.GatewayMidtrans)
	paymentStatus := "pending"

	order := orderEntity.Order{
		RegistrantID:   registrant.ID,
		OrderNumber:    orderNumber,
		Amount:         totalCost,
		Currency:       "IDR",
		PaymentGateway: &gateway,
		PaymentStatus:  paymentStatus,
	}
	err = dbTrx.GetOrderDAO().Insert(ctx, []orderEntity.Order{order})
	if err != nil {
		return nil, err
	}

	// 9. Call Payment Gateway (Midtrans) via Factory
	paymentProvider, err := s.paymentFactory.GetProvider(payment.GatewayMidtrans)
	if err != nil {
		return nil, err
	}

	// Mapping Items untuk Payment Gateway
	var paymentItems []payment.Item
	for _, t := range ticketsInvolved {
		paymentItems = append(paymentItems, payment.Item{
			ID:       string(t.ID),
			Name:     t.Title,
			Price:    t.Price,
			Quantity: 1, // Kita mapping 1-1 seperti di Laravel
		})
	}

	paymentReq := payment.CreateTransactionRequest{
		OrderID: order.OrderNumber,
		Amount:  order.Amount,
		Customer: payment.Customer{
			Name:  registrant.Name,
			Email: registrant.Email,
			Phone: registrant.Phone,
		},
		Items: paymentItems,
	}

	paymentResp, err := paymentProvider.CreateTransaction(ctx, paymentReq)
	if err != nil {
		return nil, err
	}

	// 10. Update Order dengan Token dari Gateway
	order.PaymentToken = &paymentResp.Token
	order.PaymentURL = &paymentResp.RedirectURL
	err = dbTrx.GetOrderDAO().Update(ctx, []orderEntity.Order{order})
	if err != nil {
		return nil, err
	}

	// Commit Transaction
	if err := dbTrx.GetSqlTx().Commit(); err != nil {
		s.log.Error(ctx, "registrantService.Inserts.Commit", zap.Error(err))
		return nil, err
	}

	// 11. Return Response
	return &app_registrant.RegisterResponse{
		Order: app_registrant.OrderInfo{
			OrderID:       string(order.ID),
			OrderNumber:   order.OrderNumber,
			Amount:        order.Amount,
			Currency:      order.Currency,
			PaymentStatus: order.PaymentStatus,
			PaymentToken:  paymentResp.Token,
			RedirectURL:   paymentResp.RedirectURL,
		},
		Registrant: app_registrant.RegistrantInfo{
			ID:         string(registrant.ID),
			UniqueCode: registrant.UniqueCode,
		},
	}, nil
}
