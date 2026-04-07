package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	regDao "rakit-tiket-be/internal/app/app_registrant/dao"
	"rakit-tiket-be/internal/pkg/email"
	"rakit-tiket-be/internal/pkg/payment"
	pubEntity "rakit-tiket-be/pkg/entity"
	orderEntity "rakit-tiket-be/pkg/entity/app_order"
	regEntity "rakit-tiket-be/pkg/entity/app_registrant"
	ticketEntity "rakit-tiket-be/pkg/entity/app_ticket"
	model "rakit-tiket-be/pkg/model/app_order"
	"rakit-tiket-be/pkg/util"

	"go.uber.org/zap"
)

type OrderService interface {
	HandleWebhook(ctx context.Context, gateway payment.GatewayType, payload []byte) error
	GetOrderStatus(ctx context.Context, orderNumber string) (*model.OrderStatusResponse, error)
	UpdateExpiredOrders(ctx context.Context) (int64, error)
}

type orderService struct {
	log            util.LogUtil
	sqlDB          *sql.DB
	paymentFactory *payment.PaymentFactory
	emailService   email.EmailService
}

func MakeOrderService(log util.LogUtil, sqlDB *sql.DB, paymentFactory *payment.PaymentFactory, emailService email.EmailService) OrderService {
	return orderService{
		log:            log,
		sqlDB:          sqlDB,
		paymentFactory: paymentFactory,
		emailService:   emailService,
	}
}

func (s orderService) HandleWebhook(ctx context.Context, gateway payment.GatewayType, payload []byte) error {
	provider, err := s.paymentFactory.GetProvider(gateway)
	if err != nil {
		return err
	}

	notif, err := provider.ParseWebhook(ctx, payload)
	if err != nil {
		return err
	}

	dbTrx := regDao.NewTransactionRegistrant(ctx, s.log, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	orders, err := dbTrx.GetOrderDAO().SearchForUpdate(ctx, orderEntity.OrderQuery{
		OrderNumbers: []string{notif.OrderID},
	})
	if err != nil || len(orders) == 0 {
		return fmt.Errorf("order %s tidak ditemukan", notif.OrderID)
	}
	orderData := orders[0]

	// Idempotency Check
	if orderData.PaymentStatus == "paid" || orderData.PaymentStatus == "failed" || orderData.PaymentStatus == "expired" {
		return nil
	}

	// Jika status baru adalah pending
	if notif.PaymentStatus == "pending" {
		orderData.PaymentMethod = &notif.PaymentType
		orderData.PaymentChannel = &notif.PaymentChannel
		orderData.PaymentTransactionID = &notif.TransactionID
		orderData.PaymentMetadata = &notif.RawPayload
		dbTrx.GetOrderDAO().Update(ctx, []orderEntity.Order{orderData})
		return dbTrx.GetSqlTx().Commit()
	}

	// Cari Registrant dan Attendees
	registrants, _, err := dbTrx.GetRegistrantDAO().Search(ctx, regEntity.RegistrantQuery{
		IDs: []string{string(orderData.RegistrantID)},
	})

	if err != nil || len(registrants) == 0 {
		return fmt.Errorf("registrant untuk order %s tidak ditemukan", notif.OrderID)
	}

	registrantData := registrants[0]

	attendees, err := dbTrx.GetAttendeeDAO().Search(ctx, regEntity.AttendeeQuery{
		RegistrantIDs: []string{string(registrantData.ID)},
	})

	if err != nil {
		return err
	}

	ticketQtyMap := make(map[string]int)

	if registrantData.TicketID != nil {
		ticketQtyMap[string(*registrantData.TicketID)]++
	}

	for _, att := range attendees {
		ticketQtyMap[string(att.TicketID)]++
	}

	now := time.Now()

	if notif.PaymentStatus == "paid" {
		for tID, qty := range ticketQtyMap {
			err := dbTrx.GetTicketDAO().ConfirmSold(ctx, pubEntity.UUID(tID), qty)
			if err != nil {
				return fmt.Errorf("gagal ConfirmSold tiket %s: %v", tID, err)
			}
		}
		orderData.PaymentTime = &now

		// Send Tiket
		var ticketIDs []string
		for tID := range ticketQtyMap {
			ticketIDs = append(ticketIDs, tID)
		}

		ticketMap := make(map[string]ticketEntity.Ticket)
		if len(ticketIDs) > 0 {
			tickets, err := dbTrx.GetTicketDAO().Search(ctx, ticketEntity.TicketQuery{
				IDs: ticketIDs,
			})

			if err != nil {
				// Return error agar webhook Midtrans tetap jalan, cukup log saja
				s.log.Error(ctx, "Failed to fetch ticket data for PDF", zap.Error(err))
			} else {
				for _, t := range tickets {
					ticketMap[string(t.ID)] = t
				}
			}
		}

		dynamicEvent := EventDynamicData{
			EventName:      "Rakit Tiket Event", // Fallback Default
			EventDate:      "Belum Ditentukan",
			EventTimeStart: "-",
			EventTimeEnd:   "-",
			EventLocation:  "Venue Terpilih",
		}

		// Cari Nama Event Asli NEED CEK
		_ = s.sqlDB.QueryRowContext(ctx, "SELECT name FROM events WHERE id = $1", orderData.EventID).Scan(&dynamicEvent.EventName)

		// Cari Waktu & Lokasi di tabel landing_pages
		_ = s.sqlDB.QueryRowContext(ctx, "SELECT event_date, event_time_start, event_time_end, event_location FROM landing_pages WHERE event_id = $1", orderData.EventID).
			Scan(&dynamicEvent.EventDate, &dynamicEvent.EventTimeStart, &dynamicEvent.EventTimeEnd, &dynamicEvent.EventLocation)

		attachments, err := GenerateTicketsPDF(orderData, registrantData, ticketMap, dynamicEvent)
		if err != nil {
			s.log.Error(ctx, "Failed to generate PDF tickets", zap.Error(err))
		} else {
			s.log.Info(ctx, "Successfully generated PDF tickets", zap.Int("total", len(attachments)))

			var emailAtts []email.Attachment
			for _, att := range attachments {
				emailAtts = append(emailAtts, email.Attachment{
					FileName: att.FileName,
					Data:     att.Data,
				})
			}

			go func(targetEmail, ordNum, evtName, ownerName string, atts []email.Attachment) {
				bgCtx := context.Background()

				err := s.emailService.SendTicketEmail(bgCtx, targetEmail, ordNum, evtName, ownerName, atts)
				if err != nil {
					s.log.Error(bgCtx, "Gagal mengirim email PDF asinkronus", zap.Error(err))
				} else {
					s.log.Info(bgCtx, "Email PDF E-Ticket berhasil terkirim!", zap.String("to", targetEmail))
				}
			}(registrantData.Email, orderData.OrderNumber, dynamicEvent.EventName, registrantData.Name, emailAtts)
		}

	} else if notif.PaymentStatus == "failed" || notif.PaymentStatus == "expired" {
		for tID, qty := range ticketQtyMap {
			err := dbTrx.GetTicketDAO().ReleaseBooked(ctx, pubEntity.UUID(tID), qty)
			if err != nil {
				return fmt.Errorf("gagal ReleaseBooked tiket %s: %v", tID, err)
			}
		}
	}

	// Update Status Utama
	orderData.PaymentStatus = notif.PaymentStatus
	orderData.PaymentMethod = &notif.PaymentType
	orderData.PaymentChannel = &notif.PaymentChannel
	orderData.PaymentTransactionID = &notif.TransactionID
	orderData.PaymentMetadata = &notif.RawPayload

	registrantData.Status = notif.PaymentStatus

	// Simpan Perubahan ke DB
	if err := dbTrx.GetOrderDAO().Update(ctx, []orderEntity.Order{orderData}); err != nil {
		return err
	}
	if err := dbTrx.GetRegistrantDAO().Update(ctx, []regEntity.Registrant{registrantData}); err != nil {
		return err
	}

	// Commit
	if err := dbTrx.GetSqlTx().Commit(); err != nil {
		return err
	}

	return nil
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func (s orderService) GetOrderStatus(ctx context.Context, orderNumber string) (*model.OrderStatusResponse, error) {
	dbTrx := regDao.NewTransactionRegistrant(ctx, s.log, s.sqlDB)

	orders, err := dbTrx.GetOrderDAO().Search(ctx, orderEntity.OrderQuery{
		OrderNumbers: []string{orderNumber},
	})

	if err != nil || len(orders) == 0 {
		return nil, errors.New("order not found")
	}

	orderData := orders[0]

	if orderData.PaymentStatus == "pending" && orderData.ExpiresAt != nil && time.Now().After(*orderData.ExpiresAt) {
		orderData.PaymentStatus = "expired"
		if err := dbTrx.GetOrderDAO().Update(ctx, []orderEntity.Order{orderData}); err != nil {
			s.log.Error(ctx, "Failed to update expired order", zap.Error(err))
		}
		if err := dbTrx.GetSqlTx().Commit(); err != nil {
			s.log.Error(ctx, "Failed to commit expired order update", zap.Error(err))
		}
	}

	// Ambil Registrant
	registrants, _, err := dbTrx.GetRegistrantDAO().Search(ctx, regEntity.RegistrantQuery{
		IDs: []string{string(orderData.RegistrantID)},
	})

	if err != nil || len(registrants) == 0 {
		return nil, errors.New("registrant not found")
	}

	registrantData := registrants[0]

	// Ambil Attendees
	attendees, err := dbTrx.GetAttendeeDAO().Search(ctx, regEntity.AttendeeQuery{
		RegistrantIDs: []string{string(registrantData.ID)},
	})

	if err != nil {
		return nil, err
	}

	// Kumpulkan Ticket ID & Ambil Data Tiket
	var ticketIDs []string
	if registrantData.TicketID != nil {
		ticketIDs = append(ticketIDs, string(*registrantData.TicketID))
	}

	for _, att := range attendees {
		ticketIDs = append(ticketIDs, string(att.TicketID))
	}

	ticketMap := make(map[string]ticketEntity.Ticket)

	if len(ticketIDs) > 0 {
		tickets, err := dbTrx.GetTicketDAO().Search(ctx, ticketEntity.TicketQuery{
			IDs: ticketIDs,
		})

		if err == nil {
			for _, t := range tickets {
				ticketMap[string(t.ID)] = t
			}
		}
	}

	// Format Registrant Data
	var regBirthdate *string
	if registrantData.Birthdate != nil {
		dt := registrantData.Birthdate.Format("2006-01-02")
		regBirthdate = &dt
	}

	var regTicketTitle, regTicketType *string
	if registrantData.TicketID != nil {
		if t, ok := ticketMap[string(*registrantData.TicketID)]; ok {
			regTicketTitle = &t.Title
			regTicketType = &t.Title
		}
	}

	regStatus := model.RegistrantStatus{
		Name:        registrantData.Name,
		Email:       registrantData.Email,
		Phone:       registrantData.Phone,
		Gender:      registrantData.Gender,
		Birthdate:   regBirthdate,
		TicketTitle: regTicketTitle,
		TicketType:  regTicketType,
	}

	// Format Attendees Data
	var attStatuses []model.AttendeeStatus
	for _, att := range attendees {
		var attBirthdate *string
		if att.Birthdate != nil {
			dt := att.Birthdate.Format("2006-01-02")
			attBirthdate = &dt
		}

		var attTicketTitle, attTicketType *string
		if t, ok := ticketMap[string(att.TicketID)]; ok {
			attTicketTitle = &t.Title
			attTicketType = &t.Type
		}

		attStatuses = append(attStatuses, model.AttendeeStatus{
			Name:        att.Name,
			Gender:      att.Gender,
			Birthdate:   attBirthdate,
			TicketTitle: attTicketTitle,
			TicketType:  attTicketType,
		})
	}

	// Return Response
	return &model.OrderStatusResponse{
		OrderNumber:    orderData.OrderNumber,
		PaymentMethod:  derefString(orderData.PaymentMethod),
		PaymentChannel: derefString(orderData.PaymentChannel),
		PaymentStatus:  orderData.PaymentStatus,
		Amount:         orderData.Amount,
		PaymentTime:    orderData.PaymentTime,
		Registrant:     regStatus,
		Attendees:      attStatuses,
	}, nil

}

func (s orderService) UpdateExpiredOrders(ctx context.Context) (int64, error) {
	dbTrx := regDao.NewTransactionRegistrant(ctx, s.log, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	now := time.Now()
	orders, err := dbTrx.GetOrderDAO().Search(ctx, orderEntity.OrderQuery{
		Statuses:      []string{"pending"},
		ExpiredBefore: &now,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to search expired orders: %w", err)
	}

	if len(orders) == 0 {
		return 0, nil
	}

	var expiredOrders []orderEntity.Order
	for _, order := range orders {
		order.PaymentStatus = "expired"
		expiredOrders = append(expiredOrders, order)
	}

	if err := dbTrx.GetOrderDAO().Update(ctx, expiredOrders); err != nil {
		return 0, fmt.Errorf("failed to update expired orders: %w", err)
	}

	if err := dbTrx.GetSqlTx().Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit expired orders update: %w", err)
	}

	s.log.Info(ctx, "Updated expired orders", zap.Int("count", len(expiredOrders)))
	return int64(len(expiredOrders)), nil
}
