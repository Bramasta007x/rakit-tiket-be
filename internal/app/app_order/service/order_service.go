package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	regDao "rakit-tiket-be/internal/app/app_registrant/dao"
	"rakit-tiket-be/internal/pkg/payment"
	pubEntity "rakit-tiket-be/pkg/entity"
	orderEntity "rakit-tiket-be/pkg/entity/app_order"
	regEntity "rakit-tiket-be/pkg/entity/app_registrant"
	"rakit-tiket-be/pkg/util"
)

type OrderService interface {
	HandleWebhook(ctx context.Context, gateway payment.GatewayType, payload []byte) error
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
