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
)

type OrderService interface {
	HandleWebhook(ctx context.Context, gateway payment.GatewayType, payload []byte) error
}

type orderService struct {
	sqlDB          *sql.DB
	paymentFactory *payment.PaymentFactory
}

func MakeOrderService(sqlDB *sql.DB, paymentFactory *payment.PaymentFactory) OrderService {
	return orderService{
		sqlDB:          sqlDB,
		paymentFactory: paymentFactory,
	}
}

func (s orderService) HandleWebhook(ctx context.Context, gateway payment.GatewayType, payload []byte) error {
	// 1. Parse Webhook dari Gateway yang sesuai
	provider, err := s.paymentFactory.GetProvider(gateway)
	if err != nil {
		return err
	}

	notif, err := provider.ParseWebhook(ctx, payload)
	if err != nil {
		return err
	}

	// 2. Mulai DB Transaction (Kita pinjam regDao agar bisa modif Order, Registrant, dan Ticket sekaligus)
	dbTrx := regDao.NewTransactionRegistrant(ctx, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	// 3. Cari Order berdasarkan OrderNumber
	orders, err := dbTrx.GetOrderDAO().Search(ctx, orderEntity.OrderQuery{
		OrderNumbers: []string{notif.OrderID},
	})
	if err != nil || len(orders) == 0 {
		return fmt.Errorf("order %s tidak ditemukan", notif.OrderID)
	}
	orderData := orders[0]

	// Idempotency Check: Jika status order sudah final (paid/failed/expired), abaikan request
	if orderData.PaymentStatus == "paid" || orderData.PaymentStatus == "failed" || orderData.PaymentStatus == "expired" {
		return nil // Webhook sudah diproses sebelumnya
	}

	// Jika status baru adalah pending, maka tidak ada aksi krusial, cukup update order metadata saja
	if notif.PaymentStatus == "pending" {
		orderData.PaymentMethod = &notif.PaymentType
		orderData.PaymentTransactionID = &notif.TransactionID
		orderData.PaymentMetadata = &notif.RawPayload
		dbTrx.GetOrderDAO().Update(ctx, []orderEntity.Order{orderData})
		return dbTrx.GetSqlTx().Commit()
	}

	// 4. Proses Perubahan Status Status (Berhasil atau Gagal)

	// Cari Registrant dan Attendees untuk menghitung jumlah tiket
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

	// Mapping Qty Tiket yang dibeli
	ticketQtyMap := make(map[string]int)
	if registrantData.TicketID != nil {
		ticketQtyMap[string(*registrantData.TicketID)]++
	}
	for _, att := range attendees {
		ticketQtyMap[string(att.TicketID)]++
	}

	// 5. EKSEKUSI ATOMIC TICKET BERDASARKAN STATUS
	now := time.Now()

	if notif.PaymentStatus == "paid" {
		// A. PEMBAYARAN LUNAS -> Konfirmasi Booked menjadi Sold
		for tID, qty := range ticketQtyMap {
			err := dbTrx.GetTicketDAO().ConfirmSold(ctx, pubEntity.UUID(tID), qty)
			if err != nil {
				return fmt.Errorf("gagal ConfirmSold tiket %s: %v", tID, err)
			}
		}
		orderData.PaymentTime = &now

	} else if notif.PaymentStatus == "failed" || notif.PaymentStatus == "expired" {
		// B. PEMBAYARAN GAGAL/KADALUWARSA -> Kembalikan Booked ke Available (Release)
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
