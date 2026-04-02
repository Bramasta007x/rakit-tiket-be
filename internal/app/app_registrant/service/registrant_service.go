package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"rakit-tiket-be/internal/app/app_registrant/dao"
	"rakit-tiket-be/internal/pkg/payment"
	pubEntity "rakit-tiket-be/pkg/entity"
	eventEntity "rakit-tiket-be/pkg/entity/app_event"
	orderEntity "rakit-tiket-be/pkg/entity/app_order"
	regEntity "rakit-tiket-be/pkg/entity/app_registrant"
	ticketEntity "rakit-tiket-be/pkg/entity/app_ticket"
	model "rakit-tiket-be/pkg/model/app_registrant"
	httpModel "rakit-tiket-be/pkg/model/http"
	"rakit-tiket-be/pkg/util"

	"go.uber.org/zap"
)

type RegistrantService interface {
	Register(ctx context.Context, req model.RegisterRequest) (*model.RegisterResponse, error)
	List(ctx context.Context, req model.SearchRegistrantsRequestModel) (int, model.SearchRegistrantsResponseModel)
	GetSummary(ctx context.Context) (int, model.SummaryResponseModel)
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

	// Cari Order berdasarkan OrderNumber (with row lock to prevent race condition)
	tickets, err := dbTrx.GetTicketDAO().SearchForUpdate(ctx, ticketEntity.TicketQuery{IDs: ticketIDs})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tickets: %v", err)
	}

	// Buat Map Tiket dan Pastikan Semua Tiket dari Event yang Sama
	ticketMap := make(map[string]ticketEntity.Ticket)
	var eventID pubEntity.UUID

	if len(tickets) == 0 {
		return nil, errors.New("tiket tidak ditemukan")
	}

	for i, t := range tickets {
		ticketMap[string(t.ID)] = t
		if i == 0 {
			eventID = t.EventID
		} else if t.EventID != eventID {
			return nil, errors.New("semua tiket dalam satu transaksi harus berasal dari event yang sama")
		}
	}

	if eventID == "" {
		return nil, errors.New("event_id pada tiket tidak valid")
	}

	// Ambil Konfigurasi Event
	events, err := dbTrx.GetEventDAO().Search(ctx, eventEntity.EventQuery{IDs: []string{string(eventID)}})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch event: %v", err)
	}
	if len(events) == 0 {
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

		// Eksekusi Atomic Booking!
		err := dbTrx.GetTicketDAO().BookStock(ctx, pubEntity.UUID(tID), qty)
		if err != nil {
			return nil, fmt.Errorf("stok tiket %s tidak mencukupi (habis)", ticketData.Title)
		}

		// Hitung Harga
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

	// Default fallback misal prefix kosong
	prefix := eventData.TicketPrefixCode
	if prefix == "" {
		prefix = "TKT"
	}

	// Gunakan UUID untuk memastikan unique code tidak bentrok di high concurrency
	uniqueSuffix := strings.ReplaceAll(registrantID.String(), "-", "")[:12]
	uniqueCode := fmt.Sprintf("%s-%d-%s", prefix, now.Year(), uniqueSuffix)
	orderNumber := fmt.Sprintf("%s%d-%s", prefix, now.Year(), uniqueSuffix)

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

func (s registrantService) List(ctx context.Context, req model.SearchRegistrantsRequestModel) (int, model.SearchRegistrantsResponseModel) {
	dbTrx := dao.NewTransactionRegistrant(ctx, s.log, s.sqlDB)

	query := regEntity.RegistrantQuery{
		Search:        req.Search,
		TicketTypes:   req.TicketTypes,
		PaymentStatus: req.PaymentStatus,
		DateStart:     req.DateStart,
		DateEnd:       req.DateEnd,
		SortBy:        req.SortBy,
		SortOrder:     req.SortOrder,
		DaoQuery:      req.DaoQuery,
		PagingQuery: pubEntity.PagingQuery{
			Page:    pubEntity.Page(req.Page),
			Limit:   pubEntity.Limit(req.Limit),
			NoLimit: req.NoLimit,
		},
	}

	registrants, totalCount, err := dbTrx.GetRegistrantDAO().Search(ctx, query)
	if err != nil {
		s.log.Error(ctx, "registrantService.List", zap.Error(err))
		return http.StatusInternalServerError, model.SearchRegistrantsResponseModel{}
	}

	if len(registrants) == 0 {
		return http.StatusOK, model.SearchRegistrantsResponseModel{
			HTTPResponseModel: httpModel.HTTPResponseModel{
				Code:  http.StatusOK,
				Count: 0,
			},
			Data: model.RegistrantsModel{},
		}
	}

	var regIDs []string
	for _, r := range registrants {
		regIDs = append(regIDs, string(r.ID))
	}

	orders, err := dbTrx.GetOrderDAO().Search(ctx, orderEntity.OrderQuery{
		RegistrantIDs: regIDs,
	})
	if err != nil {
		s.log.Error(ctx, "registrantService.List.GetOrders", zap.Error(err))
		return http.StatusInternalServerError, model.SearchRegistrantsResponseModel{}
	}

	orderMap := make(map[string]orderEntity.Order)
	for _, o := range orders {
		orderMap[string(o.RegistrantID)] = o
	}

	attendees, err := dbTrx.GetAttendeeDAO().Search(ctx, regEntity.AttendeeQuery{
		RegistrantIDs: regIDs,
	})
	if err != nil {
		s.log.Error(ctx, "registrantService.List.GetAttendees", zap.Error(err))
		return http.StatusInternalServerError, model.SearchRegistrantsResponseModel{}
	}

	attendeeMap := make(map[string][]regEntity.Attendee)
	for _, a := range attendees {
		attendeeMap[string(a.RegistrantID)] = append(attendeeMap[string(a.RegistrantID)], a)
	}

	ticketIDs := []string{}
	for _, r := range registrants {
		if r.TicketID != nil {
			ticketIDs = append(ticketIDs, string(*r.TicketID))
		}
	}
	for _, atts := range attendeeMap {
		for _, a := range atts {
			ticketIDs = append(ticketIDs, string(a.TicketID))
		}
	}

	tickets, err := dbTrx.GetTicketDAO().Search(ctx, ticketEntity.TicketQuery{
		IDs: ticketIDs,
	})
	if err != nil {
		s.log.Error(ctx, "registrantService.List.GetTickets", zap.Error(err))
		return http.StatusInternalServerError, model.SearchRegistrantsResponseModel{}
	}

	ticketMap := make(map[string]ticketEntity.Ticket)
	for _, t := range tickets {
		ticketMap[string(t.ID)] = t
	}

	baseURL := "/api/v1/admin/ticket/"

	page := int(query.PagingQuery.Page)
	limit := int(query.PagingQuery.Limit)

	return model.MakeSearchRegistrantsResponseModel(http.StatusOK, totalCount, registrants, orderMap, ticketMap, attendeeMap, baseURL, page, limit)
}

func (s registrantService) GetSummary(ctx context.Context) (int, model.SummaryResponseModel) {
	dbTrx := dao.NewTransactionRegistrant(ctx, s.log, s.sqlDB)

	registrants, _, err := dbTrx.GetRegistrantDAO().Search(ctx, regEntity.RegistrantQuery{
		DaoQuery: pubEntity.DaoQuery{
			Deleted: []bool{false},
		},
	})
	if err != nil {
		s.log.Error(ctx, "registrantService.GetSummary.SearchRegistrants", zap.Error(err))
		return http.StatusInternalServerError, model.SummaryResponseModel{}
	}

	var regIDs []string
	for _, r := range registrants {
		regIDs = append(regIDs, string(r.ID))
	}

	orders, err := dbTrx.GetOrderDAO().Search(ctx, orderEntity.OrderQuery{
		RegistrantIDs: regIDs,
	})
	if err != nil {
		s.log.Error(ctx, "registrantService.GetSummary.GetOrders", zap.Error(err))
		return http.StatusInternalServerError, model.SummaryResponseModel{}
	}

	attendees, err := dbTrx.GetAttendeeDAO().Search(ctx, regEntity.AttendeeQuery{
		RegistrantIDs: regIDs,
	})
	if err != nil {
		s.log.Error(ctx, "registrantService.GetSummary.GetAttendees", zap.Error(err))
		return http.StatusInternalServerError, model.SummaryResponseModel{}
	}

	summary := model.SummaryData{
		TotalRegistrants:   len(registrants),
		TotalAttendees:     len(attendees),
		TotalTickets:       0,
		TotalRevenue:       0,
		PaidRegistrants:    0,
		PendingRegistrants: 0,
		FailedRegistrants:  0,
	}

	for _, r := range registrants {
		summary.TotalTickets += r.TotalTickets
	}

	for _, o := range orders {
		switch o.PaymentStatus {
		case "paid":
			summary.PaidRegistrants++
			summary.TotalRevenue += o.Amount
		case "pending":
			summary.PendingRegistrants++
		case "failed":
			summary.FailedRegistrants++
		}
	}

	return model.MakeSummaryResponseModel(http.StatusOK, summary)
}
