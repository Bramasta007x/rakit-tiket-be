package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
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
	"rakit-tiket-be/pkg/util"

	"go.uber.org/zap"
)

type RegistrantService interface {
	Register(ctx context.Context, req model.RegisterRequest) (*model.RegisterResponse, error)
	List(ctx context.Context, filter model.ListFilter) (*model.ListResponse, error)
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

func (s registrantService) List(ctx context.Context, filter model.ListFilter) (*model.ListResponse, error) {
	dbTrx := dao.NewTransactionRegistrant(ctx, s.log, s.sqlDB)

	perPage := 25
	if filter.PerPage > 0 {
		perPage = filter.PerPage
	}

	page := 1
	if filter.Page > 0 {
		page = filter.Page
	}

	var args []interface{}
	argIndex := 1

	var whereConditions []string
	whereConditions = append(whereConditions, "r.deleted = false")

	if filter.Search != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("(r.name ILIKE $%d OR r.email ILIKE $%d)", argIndex, argIndex))
		args = append(args, "%"+filter.Search+"%")
		argIndex++
	}

	if len(filter.PaymentStatus) > 0 {
		placeholders := make([]string, len(filter.PaymentStatus))
		for i, status := range filter.PaymentStatus {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, status)
			argIndex++
		}
		whereConditions = append(whereConditions, fmt.Sprintf("o.payment_status IN (%s)", strings.Join(placeholders, ", ")))
	}

	if filter.DateStart != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("DATE(r.created_at) >= $%d", argIndex))
		args = append(args, filter.DateStart)
		argIndex++
	}

	if filter.DateEnd != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("DATE(r.created_at) <= $%d", argIndex))
		args = append(args, filter.DateEnd)
		argIndex++
	}

	if len(filter.TicketType) > 0 {
		placeholders := make([]string, len(filter.TicketType))
		for i, ticketType := range filter.TicketType {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, ticketType)
			argIndex++
		}
		whereConditions = append(whereConditions, fmt.Sprintf("(rt.type IN (%s) OR at.type IN (%s))", strings.Join(placeholders, ", "), strings.Join(placeholders, ", ")))
	}

	whereClause := strings.Join(whereConditions, " AND ")

	sortBy := "r.created_at"
	switch filter.SortBy {
	case "name":
		sortBy = "r.name"
	case "email":
		sortBy = "r.email"
	case "created_at":
		sortBy = "r.created_at"
	case "order_number":
		sortBy = "o.order_number"
	}

	sortOrder := "DESC"
	if filter.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	countQuery := fmt.Sprintf(`
		SELECT COUNT(DISTINCT r.id)
		FROM registrants r
		LEFT JOIN orders o ON o.registrant_id = r.id AND o.deleted = false
		LEFT JOIN tickets rt ON rt.id = r.ticket_id AND rt.deleted = false
		LEFT JOIN attendees a ON a.registrant_id = r.id AND a.deleted = false
		LEFT JOIN tickets at ON at.id = a.ticket_id AND at.deleted = false
		WHERE %s`, whereClause)

	s.log.Debug(ctx, "registrantService.List.Count",
		zap.String("SQL", countQuery),
		zap.Any("Params", args),
	)

	var totalCount int
	err := dbTrx.GetSqlDB().QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		s.log.Error(ctx, "registrantService.List.Count",
			zap.String("SQL", countQuery),
			zap.Any("Params", args),
			zap.Error(err),
		)
		return nil, err
	}

	dataQuery := fmt.Sprintf(`
		SELECT 
			r.id AS registrant_id,
			r.name AS registrant_name,
			r.email AS registrant_email,
			r.phone AS registrant_phone,
			r.gender AS registrant_gender,
			r.birthdate AS registrant_birthdate,
			r.ticket_id AS registrant_ticket_id,
			r.created_at AS registrant_created_at,
			rt.title AS registrant_ticket_title,
			rt.type AS registrant_ticket_type,
			
			o.id AS order_id,
			o.order_number AS order_number,
			o.amount AS order_amount,
			o.payment_status AS order_payment_status,
			o.payment_method AS order_payment_method,
			o.payment_channel AS order_payment_channel,
			o.payment_time AS order_payment_time,
			
			a.id AS attendee_id,
			a.name AS attendee_name,
			a.gender AS attendee_gender,
			a.birthdate AS attendee_birthdate,
			a.ticket_id AS attendee_ticket_id,
			at.title AS attendee_ticket_title,
			at.type AS attendee_ticket_type,
			
			COALESCE(r.created_at, o.created_at) AS created_at
		FROM registrants r
		LEFT JOIN orders o ON o.registrant_id = r.id AND o.deleted = false
		LEFT JOIN tickets rt ON rt.id = r.ticket_id AND rt.deleted = false
		LEFT JOIN attendees a ON a.registrant_id = r.id AND a.deleted = false
		LEFT JOIN tickets at ON at.id = a.ticket_id AND at.deleted = false
		WHERE %s
		ORDER BY %s %s
		LIMIT %d OFFSET %d`, whereClause, sortBy, sortOrder, perPage, (page-1)*perPage)

	s.log.Debug(ctx, "registrantService.List.Data",
		zap.String("SQL", dataQuery),
		zap.Any("Params", args),
	)

	rows, err := dbTrx.GetSqlDB().QueryContext(ctx, dataQuery, args...)
	if err != nil {
		s.log.Error(ctx, "registrantService.List.Data",
			zap.String("SQL", dataQuery),
			zap.Any("Params", args),
			zap.Error(err),
		)
		return nil, err
	}
	defer rows.Close()

	type rawRow struct {
		RegistrantID          string
		RegistrantName        string
		RegistrantEmail       string
		RegistrantPhone       string
		RegistrantGender      sql.NullString
		RegistrantBirthdate   sql.NullTime
		RegistrantTicketID    sql.NullString
		RegistrantCreatedAt   sql.NullTime
		RegistrantTicketTitle sql.NullString
		RegistrantTicketType  sql.NullString
		OrderID               sql.NullString
		OrderNumber           sql.NullString
		OrderAmount           sql.NullFloat64
		OrderPaymentStatus    sql.NullString
		OrderPaymentMethod    sql.NullString
		OrderPaymentChannel   sql.NullString
		OrderPaymentTime      sql.NullTime
		AttendeeID            sql.NullString
		AttendeeName          sql.NullString
		AttendeeGender        sql.NullString
		AttendeeBirthdate     sql.NullTime
		AttendeeTicketID      sql.NullString
		AttendeeTicketTitle   sql.NullString
		AttendeeTicketType    sql.NullString
		CreatedAt             sql.NullTime
	}

	var rawData []rawRow
	for rows.Next() {
		var r rawRow
		if err := rows.Scan(
			&r.RegistrantID, &r.RegistrantName, &r.RegistrantEmail, &r.RegistrantPhone,
			&r.RegistrantGender, &r.RegistrantBirthdate, &r.RegistrantTicketID,
			&r.RegistrantCreatedAt, &r.RegistrantTicketTitle, &r.RegistrantTicketType,
			&r.OrderID, &r.OrderNumber, &r.OrderAmount, &r.OrderPaymentStatus,
			&r.OrderPaymentMethod, &r.OrderPaymentChannel, &r.OrderPaymentTime,
			&r.AttendeeID, &r.AttendeeName, &r.AttendeeGender, &r.AttendeeBirthdate,
			&r.AttendeeTicketID, &r.AttendeeTicketTitle, &r.AttendeeTicketType,
			&r.CreatedAt,
		); err != nil {
			s.log.Error(ctx, "registrantService.List.Scan", zap.Error(err))
			return nil, err
		}
		rawData = append(rawData, r)
	}

	type attendeeData struct {
		TicketTitle string
		TicketType  string
		Name        string
		Gender      string
		Birthdate   *string
		OrderNum    string
	}

	grouped := make(map[string]*model.ListItem)
	attendeesMap := make(map[string][]attendeeData)

	for _, r := range rawData {
		if _, exists := grouped[r.RegistrantID]; !exists {
			var birthdateStr *string
			if r.RegistrantBirthdate.Valid {
				t := r.RegistrantBirthdate.Time.Format("2006-01-02")
				birthdateStr = &t
			}

			var regGender string
			if r.RegistrantGender.Valid {
				regGender = r.RegistrantGender.String
			}

			ticketTitle := "-"
			ticketType := "-"
			if r.RegistrantTicketTitle.Valid {
				ticketTitle = r.RegistrantTicketTitle.String
			}
			if r.RegistrantTicketType.Valid {
				ticketType = r.RegistrantTicketType.String
			}

			var paymentStatus, paymentMethod, paymentChannel string
			var orderID, orderNumber string
			var orderTotal float64
			var paymentTime *time.Time

			if r.OrderID.Valid {
				orderID = r.OrderID.String
			}
			if r.OrderNumber.Valid {
				orderNumber = r.OrderNumber.String
			}
			if r.OrderPaymentStatus.Valid {
				paymentStatus = r.OrderPaymentStatus.String
			}
			if r.OrderPaymentMethod.Valid {
				paymentMethod = r.OrderPaymentMethod.String
			}
			if r.OrderPaymentChannel.Valid {
				paymentChannel = r.OrderPaymentChannel.String
			}
			if r.OrderAmount.Valid {
				orderTotal = r.OrderAmount.Float64
			}
			if r.OrderPaymentTime.Valid {
				paymentTime = &r.OrderPaymentTime.Time
			}

			var formattedPaymentMethod string
			if paymentChannel != "" {
				formattedPaymentMethod = strings.Title(strings.ReplaceAll(strings.ReplaceAll(paymentChannel, "_", " "), "-", " "))
			} else {
				formattedPaymentMethod = "-"
			}

			item := &model.ListItem{
				UniqueID:    r.RegistrantID,
				OrderNumber: orderNumber,
				Registrant: model.RegistrantInfoDetail{
					Name:        r.RegistrantName,
					TicketTitle: ticketTitle,
					TicketType:  ticketType,
					Email:       r.RegistrantEmail,
					Phone:       r.RegistrantPhone,
					Gender:      regGender,
					Birthdate:   birthdateStr,
					ETicket:     nil,
				},
				Payment: model.PaymentInfo{
					ID:            orderID,
					Status:        paymentStatus,
					Method:        paymentMethod,
					PaymentMethod: formattedPaymentMethod,
					Time:          paymentTime,
					Total:         orderTotal,
				},
				Attendees:    []model.AttendeeInfo{},
				TicketTypes:  []string{},
				TicketTitles: []string{},
			}

			if ticketType != "-" {
				item.TicketTypes = append(item.TicketTypes, ticketType)
			}
			if ticketTitle != "-" {
				item.TicketTitles = append(item.TicketTitles, ticketTitle)
			}

			grouped[r.RegistrantID] = item
		}

		if r.AttendeeID.Valid {
			var attGender string
			if r.AttendeeGender.Valid {
				attGender = r.AttendeeGender.String
			}

			var attBirthdate *string
			if r.AttendeeBirthdate.Valid {
				t := r.AttendeeBirthdate.Time.Format("2006-01-02")
				attBirthdate = &t
			}

			attTicketTitle := "-"
			attTicketType := "-"
			if r.AttendeeTicketTitle.Valid {
				attTicketTitle = r.AttendeeTicketTitle.String
			}
			if r.AttendeeTicketType.Valid {
				attTicketType = r.AttendeeTicketType.String
			}

			attendeesMap[r.RegistrantID] = append(attendeesMap[r.RegistrantID], attendeeData{
				TicketTitle: attTicketTitle,
				TicketType:  attTicketType,
				Name:        r.AttendeeName.String,
				Gender:      attGender,
				Birthdate:   attBirthdate,
				OrderNum:    r.OrderNumber.String,
			})

			if attTicketType != "-" {
				grouped[r.RegistrantID].TicketTypes = append(grouped[r.RegistrantID].TicketTypes, attTicketType)
			}
			if attTicketTitle != "-" {
				grouped[r.RegistrantID].TicketTitles = append(grouped[r.RegistrantID].TicketTitles, attTicketTitle)
			}
		}
	}

	baseURL := "http://localhost:8000/storage/tickets/"
	makePdfName := func(orderNumber, name string) string {
		safeName := strings.ReplaceAll(name, " ", "_")
		return fmt.Sprintf("E-Voucher-%s-%s.pdf", orderNumber, safeName)
	}

	var items []model.ListItem
	for regID, item := range grouped {
		item.TotalTickets = 1 + len(attendeesMap[regID])

		if item.OrderNumber != "" {
			item.Registrant.ETicket = strPtr(baseURL + makePdfName(item.OrderNumber, item.Registrant.Name))
		}

		for _, att := range attendeesMap[regID] {
			var eticket *string
			if item.OrderNumber != "" {
				eticket = strPtr(baseURL + makePdfName(att.OrderNum, att.Name))
			}

			item.Attendees = append(item.Attendees, model.AttendeeInfo{
				TicketTitle: att.TicketTitle,
				TicketType:  att.TicketType,
				Name:        att.Name,
				Gender:      att.Gender,
				Birthdate:   att.Birthdate,
				ETicket:     eticket,
			})
		}

		items = append(items, *item)
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(perPage)))

	return &model.ListResponse{
		Data:       items,
		Total:      totalCount,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

func strPtr(s string) *string {
	return &s
}
