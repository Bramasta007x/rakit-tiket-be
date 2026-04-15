package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	paymentSvc "rakit-tiket-be/internal/app/app_payment/service"
	"rakit-tiket-be/internal/app/app_registrant/dao"
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
	GetDashboard(ctx context.Context, req model.DashboardRequestModel) (int, model.DashboardResponseModel)
}

type PaymentConfigProvider interface {
	GetActivePaymentOptions(ctx context.Context) ([]paymentSvc.PaymentOption, error)
}

type registrantService struct {
	log                   util.LogUtil
	sqlDB                 *sql.DB
	checkoutInitiator     paymentSvc.CheckoutInitiator
	paymentConfigProvider PaymentConfigProvider
}

func MakeRegistrantService(
	log util.LogUtil,
	sqlDB *sql.DB,
	checkoutInitiator paymentSvc.CheckoutInitiator,
	paymentConfigProvider PaymentConfigProvider,
) RegistrantService {
	return registrantService{
		log:                   log,
		sqlDB:                 sqlDB,
		checkoutInitiator:     checkoutInitiator,
		paymentConfigProvider: paymentConfigProvider,
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

	// ATOMIC BOOKING STOCK
	var totalCost float64

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

	// Insert Data Order (Checkout akan dilakukan terpisah)
	expiresAt := now.Add(15 * time.Minute)

	order := orderEntity.Order{
		ID:            orderID,
		EventID:       eventID,
		RegistrantID:  registrantID,
		OrderNumber:   orderNumber,
		Amount:        totalCost,
		Currency:      "IDR",
		PaymentStatus: orderEntity.OrderStatusPending,
		ExpiresAt:     &expiresAt,
	}
	order.CreatedAt = now

	if err := dbTrx.GetOrderDAO().Insert(ctx, []orderEntity.Order{order}); err != nil {
		return nil, err
	}

	if err := dbTrx.GetSqlTx().Commit(); err != nil {
		return nil, err
	}

	response := &model.RegisterResponse{
		Order: model.OrderInfo{
			OrderID:       string(order.ID),
			OrderNumber:   order.OrderNumber,
			Amount:        order.Amount,
			Currency:      order.Currency,
			PaymentStatus: order.PaymentStatus,
			ExpiresAt:     order.ExpiresAt,
		},
		Registrant: model.RegistrantInfo{
			ID:         string(registrant.ID),
			UniqueCode: registrant.UniqueCode,
		},
	}

	paymentOptions, err := s.paymentConfigProvider.GetActivePaymentOptions(ctx)
	if err != nil {
		s.log.Error(ctx, "Register.GetActivePaymentOptions", zap.Error(err))
		return response, nil
	}

	if len(paymentOptions) == 0 {
		return response, nil
	}

	response.PaymentOptions = paymentOptions

	if len(paymentOptions) == 1 {
		option := paymentOptions[0]

		if option.Type == "GATEWAY" {
			initiateResult, err := s.checkoutInitiator.InitiateGatewayPayment(ctx, &order)
			if err != nil {
				s.log.Error(ctx, "Register.InitiateGatewayPayment", zap.Error(err))
				return response, nil
			}

			response.PaymentInfo = &paymentSvc.RegisterPaymentInfo{
				PaymentType:  initiateResult.PaymentType,
				PaymentURL:   initiateResult.PaymentInfo.PaymentURL,
				PaymentToken: initiateResult.PaymentInfo.PaymentToken,
			}

			response.Order.PaymentStatus = order.PaymentStatus
		}
	}

	return response, nil
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

	summary.TotalRegistrantsAttendees = summary.TotalRegistrants + summary.TotalAttendees

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

func (s registrantService) GetDashboard(ctx context.Context, req model.DashboardRequestModel) (int, model.DashboardResponseModel) {
	dbTrx := dao.NewTransactionRegistrant(ctx, s.log, s.sqlDB)

	days := 7
	if req.Days > 0 {
		days = req.Days
	}
	limit := 10
	if req.RecentLimit > 0 {
		limit = req.RecentLimit
	}

	summary := s.buildDashboardSummary(ctx, dbTrx)
	dailySales := s.buildDailySalesTrend(ctx, dbTrx, days)
	ticketDist := s.buildTicketDistribution(ctx, dbTrx)
	recentTx := s.buildRecentTransactions(ctx, dbTrx, limit)

	dashboardData := regEntity.DashboardData{
		Summary:             summary,
		DailySalesTrend:     dailySales,
		TicketDistributions: ticketDist,
		RecentTransactions:  recentTx,
	}

	return model.MakeDashboardResponseModel(http.StatusOK, dashboardData)
}

func (s registrantService) buildDashboardSummary(ctx context.Context, dbTrx dao.DBTransaction) regEntity.DashboardSummary {
	now := time.Now()
	thisMonthStart := now.AddDate(0, 0, -30)
	lastMonthStart := now.AddDate(0, -1, -30)

	allRegistrants, _, _ := dbTrx.GetRegistrantDAO().Search(ctx, regEntity.RegistrantQuery{
		DaoQuery: pubEntity.DaoQuery{Deleted: []bool{false}},
	})
	thisMonthRegistrants, _, _ := dbTrx.GetRegistrantDAO().Search(ctx, regEntity.RegistrantQuery{
		DaoQuery: pubEntity.DaoQuery{Deleted: []bool{false}},
	})

	allOrders, _ := dbTrx.GetOrderDAO().Search(ctx, orderEntity.OrderQuery{})
	thisMonthOrders, _ := dbTrx.GetOrderDAO().Search(ctx, orderEntity.OrderQuery{})

	paidRegMap := make(map[string]bool)
	for _, o := range allOrders {
		if o.PaymentStatus == "paid" {
			paidRegMap[string(o.RegistrantID)] = true
		}
	}

	thisMonthPaidRegMap := make(map[string]bool)
	for _, o := range thisMonthOrders {
		if o.PaymentStatus == "paid" {
			thisMonthPaidRegMap[string(o.RegistrantID)] = true
		}
	}

	var thisTicketsSold, lastTicketsSold int
	var thisRegistrants, lastRegistrants int
	var thisRevenue, lastRevenue float64
	var activeEventIDs []string

	eventIDMap := make(map[string]bool)

	for _, r := range allRegistrants {
		if _, exists := eventIDMap[string(r.EventID)]; !exists {
			eventIDMap[string(r.EventID)] = true
			activeEventIDs = append(activeEventIDs, string(r.EventID))
		}
	}

	for _, r := range thisMonthRegistrants {
		if (r.CreatedAt.After(thisMonthStart) || r.CreatedAt.Equal(thisMonthStart)) && thisMonthPaidRegMap[string(r.ID)] {
			thisRegistrants++
			thisTicketsSold += r.TotalTickets
		}
	}

	for _, r := range allRegistrants {
		if r.CreatedAt.After(lastMonthStart) && r.CreatedAt.Before(thisMonthStart) && paidRegMap[string(r.ID)] {
			lastRegistrants++
			lastTicketsSold += r.TotalTickets
		}
	}

	for _, o := range thisMonthOrders {
		if o.CreatedAt.After(thisMonthStart) || o.CreatedAt.Equal(thisMonthStart) {
			if o.PaymentStatus == "paid" {
				thisRevenue += o.Amount
			}
		}
	}

	for _, o := range allOrders {
		if o.CreatedAt.After(lastMonthStart) && o.CreatedAt.Before(thisMonthStart) {
			if o.PaymentStatus == "paid" {
				lastRevenue += o.Amount
			}
		}
	}

	calcChange := func(current, previous int) float64 {
		if previous == 0 {
			return 0
		}
		return float64(current-previous) / float64(previous) * 100
	}

	return regEntity.DashboardSummary{
		TotalTicketsSold:  thisTicketsSold,
		TicketsSoldChange: calcChange(thisTicketsSold, lastTicketsSold),
		TotalRegistrants:  thisRegistrants,
		RegistrantsChange: calcChange(thisRegistrants, lastRegistrants),
		TotalRevenue:      thisRevenue,
		RevenueChange:     calcChange(int(thisRevenue), int(lastRevenue)),
		ActiveEvents:      len(activeEventIDs),
	}
}

func (s registrantService) buildDailySalesTrend(ctx context.Context, dbTrx dao.DBTransaction, days int) regEntity.DailySalesTrend {
	now := time.Now()
	startDate := now.AddDate(0, 0, -days)

	orders, _ := dbTrx.GetOrderDAO().Search(ctx, orderEntity.OrderQuery{})
	registrants, _, _ := dbTrx.GetRegistrantDAO().Search(ctx, regEntity.RegistrantQuery{
		DaoQuery: pubEntity.DaoQuery{Deleted: []bool{false}},
	})

	regMap := make(map[string]regEntity.Registrant)
	for _, r := range registrants {
		regMap[string(r.ID)] = r
	}

	dailyMap := make(map[string]regEntity.DailySales)
	for i := days - 1; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")
		dailyMap[dateStr] = regEntity.DailySales{Date: dateStr, TicketsSold: 0, Revenue: 0}
	}

	for _, o := range orders {
		if o.CreatedAt.After(startDate) || o.CreatedAt.Equal(startDate) {
			dateStr := o.CreatedAt.Format("2006-01-02")
			if ds, exists := dailyMap[dateStr]; exists {
				if r, ok := regMap[string(o.RegistrantID)]; ok {
					ds.TicketsSold += r.TotalTickets
					if o.PaymentStatus == "paid" {
						ds.Revenue += o.Amount
					}
					dailyMap[dateStr] = ds
				}
			}
		}
	}

	var trend regEntity.DailySalesTrend
	for i := days - 1; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")
		trend = append(trend, dailyMap[dateStr])
	}

	return trend
}

func (s registrantService) buildTicketDistribution(ctx context.Context, dbTrx dao.DBTransaction) regEntity.TicketDistributions {
	tickets, _ := dbTrx.GetTicketDAO().Search(ctx, ticketEntity.TicketQuery{})
	registrants, _, _ := dbTrx.GetRegistrantDAO().Search(ctx, regEntity.RegistrantQuery{
		DaoQuery: pubEntity.DaoQuery{Deleted: []bool{false}},
	})

	typeCount := make(map[string]struct {
		sold     int
		capacity int
	})
	totalSold := 0

	for _, r := range registrants {
		if r.TicketID != nil {
			ticketID := string(*r.TicketID)
			if _, exists := typeCount[ticketID]; !exists {
				for _, t := range tickets {
					if string(t.ID) == ticketID {
						typeCount[ticketID] = struct {
							sold     int
							capacity int
						}{sold: 0, capacity: t.Total}
						break
					}
				}
			}
			count := typeCount[ticketID]
			count.sold += r.TotalTickets
			totalSold += r.TotalTickets
			typeCount[ticketID] = count
		}
	}

	var distributions regEntity.TicketDistributions
	ticketTypeMap := make(map[string]string)

	for _, t := range tickets {
		ticketTypeMap[string(t.ID)] = t.Type
	}

	for ticketID, data := range typeCount {
		ticketType := ticketTypeMap[ticketID]
		if ticketType == "" {
			ticketType = "UNKNOWN"
		}

		percentage := 0.0
		if totalSold > 0 {
			percentage = float64(data.sold) / float64(totalSold) * 100
		}

		distributions = append(distributions, regEntity.TicketDistribution{
			TicketType:    ticketType,
			TicketsSold:   data.sold,
			TotalCapacity: data.capacity,
			Percentage:    percentage,
		})
	}

	return distributions
}

func (s registrantService) buildRecentTransactions(ctx context.Context, dbTrx dao.DBTransaction, limit int) regEntity.RecentTransactions {
	orders, _ := dbTrx.GetOrderDAO().Search(ctx, orderEntity.OrderQuery{})
	registrants, _, _ := dbTrx.GetRegistrantDAO().Search(ctx, regEntity.RegistrantQuery{
		DaoQuery: pubEntity.DaoQuery{Deleted: []bool{false}},
	})
	tickets, _ := dbTrx.GetTicketDAO().Search(ctx, ticketEntity.TicketQuery{})

	regMap := make(map[string]regEntity.Registrant)
	for _, r := range registrants {
		regMap[string(r.ID)] = r
	}

	ticketMap := make(map[string]ticketEntity.Ticket)
	for _, t := range tickets {
		ticketMap[string(t.ID)] = t
	}

	type transactionWithTime struct {
		order      orderEntity.Order
		registrant regEntity.Registrant
	}

	var txs []transactionWithTime
	for _, o := range orders {
		if r, ok := regMap[string(o.RegistrantID)]; ok {
			txs = append(txs, transactionWithTime{order: o, registrant: r})
		}
	}

	sort.Slice(txs, func(i, j int) bool {
		return txs[i].order.CreatedAt.After(txs[j].order.CreatedAt)
	})

	if len(txs) > limit {
		txs = txs[:limit]
	}

	now := time.Now()
	var recentTx regEntity.RecentTransactions
	for _, tx := range txs {
		ticketType := "-"
		if tx.registrant.TicketID != nil {
			if t, ok := ticketMap[string(*tx.registrant.TicketID)]; ok {
				ticketType = t.Type
			}
		}

		recentTx = append(recentTx, regEntity.RecentTransaction{
			ID:         string(tx.order.ID),
			BuyerName:  tx.registrant.Name,
			TicketType: ticketType,
			Quantity:   tx.registrant.TotalTickets,
			Amount:     tx.order.Amount,
			Status:     tx.order.PaymentStatus,
			TimeAgo:    timeSince(tx.order.CreatedAt, now),
		})
	}

	return recentTx
}

func timeSince(t, now time.Time) string {
	diff := now.Sub(t)
	switch {
	case diff < time.Minute:
		return "baru saja"
	case diff < time.Hour:
		return fmt.Sprintf("%d menit lalu", int(diff.Minutes()))
	case diff < 24*time.Hour:
		return fmt.Sprintf("%d jam lalu", int(diff.Hours()))
	case diff < 7*24*time.Hour:
		return fmt.Sprintf("%d hari lalu", int(diff.Hours()/24))
	default:
		return t.Format("02 Jan 2006")
	}
}
