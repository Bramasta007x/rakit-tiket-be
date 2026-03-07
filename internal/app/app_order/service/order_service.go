package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	regDao "rakit-tiket-be/internal/app/app_registrant/dao"
	"rakit-tiket-be/internal/pkg/payment"
	pubEntity "rakit-tiket-be/pkg/entity"
	orderEntity "rakit-tiket-be/pkg/entity/app_order"
	regEntity "rakit-tiket-be/pkg/entity/app_registrant"
	ticketEntity "rakit-tiket-be/pkg/entity/app_ticket"
	model "rakit-tiket-be/pkg/model/app_order"
	"rakit-tiket-be/pkg/util"
)

type OrderService interface {
	HandleWebhook(ctx context.Context, gateway payment.GatewayType, payload []byte) error
	GetOrderStatus(ctx context.Context, orderNumber string) (*model.OrderStatusResponse, error)
}

type orderService struct {
	log            util.LogUtil
	sqlDB          *sql.DB
	paymentFactory *payment.PaymentFactory
}

func MakeOrderService(log util.LogUtil, sqlDB *sql.DB, paymentFactory *payment.PaymentFactory) OrderService {
	return orderService{
		log:            log,
		sqlDB:          sqlDB,
		paymentFactory: paymentFactory,
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

	orders, err := dbTrx.GetOrderDAO().Search(ctx, orderEntity.OrderQuery{
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
	registrants, err := dbTrx.GetRegistrantDAO().Search(ctx, regEntity.RegistrantQuery{
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

	} else if notif.PaymentStatus == "failed" || notif.PaymentStatus == "expired" {
		for tID, qty := range ticketQtyMap {
			err := dbTrx.GetTicketDAO().ReleaseBooked(ctx, pubEntity.UUID(tID), qty)
			if err != nil {
				return fmt.Errorf("gagal ReleaseBooked tiket %s: %v", tID, err)
			}
		}
	}

	// 6. Update Status Utama
	orderData.PaymentStatus = notif.PaymentStatus
	orderData.PaymentMethod = &notif.PaymentType
	orderData.PaymentTransactionID = &notif.TransactionID
	orderData.PaymentMetadata = &notif.RawPayload

	registrantData.Status = notif.PaymentStatus

	// 7. Simpan Perubahan ke DB
	if err := dbTrx.GetOrderDAO().Update(ctx, []orderEntity.Order{orderData}); err != nil {
		return err
	}
	if err := dbTrx.GetRegistrantDAO().Update(ctx, []regEntity.Registrant{registrantData}); err != nil {
		return err
	}

	// 8. Selesaikan Transaksi
	if err := dbTrx.GetSqlTx().Commit(); err != nil {
		return err
	}

	return nil
}

func (s orderService) GetOrderStatus(ctx context.Context, orderNumber string) (*model.OrderStatusResponse, error) {
	dbTrx := regDao.NewTransactionRegistrant(ctx, s.log, s.sqlDB)

	// Ambil Order
	orders, err := dbTrx.GetOrderDAO().Search(ctx, orderEntity.OrderQuery{
		OrderNumbers: []string{orderNumber},
	})

	if err != nil || len(orders) == 0 {
		return nil, errors.New("order not found")
	}

	orderData := orders[0]

	// Ambil Registrant
	registrants, err := dbTrx.GetRegistrantDAO().Search(ctx, regEntity.RegistrantQuery{
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
		OrderNumber:   orderData.OrderNumber,
		PaymentMethod: *orderData.PaymentMethod,
		PaymentStatus: orderData.PaymentStatus,
		Amount:        orderData.Amount,
		PaymentTime:   orderData.PaymentTime,
		Registrant:    regStatus,
		Attendees:     attStatuses,
	}, nil

}
