package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	orderSvc "rakit-tiket-be/internal/app/app_order/service"
	"rakit-tiket-be/internal/app/app_payment/dao"
	"rakit-tiket-be/internal/pkg/email"
	pubEntity "rakit-tiket-be/pkg/entity"
	orderEntity "rakit-tiket-be/pkg/entity/app_order"
	appPayment "rakit-tiket-be/pkg/entity/app_payment"
	regEntity "rakit-tiket-be/pkg/entity/app_registrant"
	ticketEntity "rakit-tiket-be/pkg/entity/app_ticket"
	"rakit-tiket-be/pkg/util"

	"go.uber.org/zap"
)

var (
	ErrManualTransferNotFound = errors.New("manual transfer not found")
	ErrOrderNotFound          = errors.New("order not found")
	ErrTransferAlreadyExists  = errors.New("transfer proof already submitted for this order")
	ErrInvalidStatus          = errors.New("invalid transfer status")
)

type ManualTransferService interface {
	SubmitTransferProof(ctx context.Context, req SubmitTransferProofRequest) (*appPayment.ManualTransfer, error)
	GetPendingTransfers(ctx context.Context) ([]ManualTransferWithDetails, error)
	GetTransferByOrderID(ctx context.Context, orderID string) (*appPayment.ManualTransfer, error)
	ApproveTransfer(ctx context.Context, transferID string, adminID string, notes string) error
	RejectTransfer(ctx context.Context, transferID string, adminID string, notes string) error
	CancelTransfer(ctx context.Context, transferID string, adminID string, notes string) error
}

type manualTransferService struct {
	log          util.LogUtil
	sqlDB        *sql.DB
	emailService email.EmailService
}

func MakeManualTransferService(log util.LogUtil, sqlDB *sql.DB, emailService email.EmailService) ManualTransferService {
	return &manualTransferService{
		log:          log,
		sqlDB:        sqlDB,
		emailService: emailService,
	}
}

func (s *manualTransferService) SubmitTransferProof(ctx context.Context, req SubmitTransferProofRequest) (*appPayment.ManualTransfer, error) {
	dbTrx := dao.NewTransactionPayment(ctx, s.log, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	existingTransfer, err := dbTrx.GetManualTransferDAO().GetByOrderID(ctx, req.OrderID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing transfer: %w", err)
	}
	if existingTransfer != nil {
		return nil, ErrTransferAlreadyExists
	}

	orders, err := dbTrx.GetOrderDAO().Search(ctx, orderEntity.OrderQuery{
		IDs: []string{string(req.OrderID)},
	})
	if err != nil || len(orders) == 0 {
		return nil, ErrOrderNotFound
	}
	order := orders[0]

	transfer := appPayment.ManualTransfer{
		ID:                    pubEntity.MakeUUID("MT", string(req.OrderID), time.Now().String()),
		OrderID:               req.OrderID,
		BankAccountID:         req.BankAccountID,
		TransferAmount:        order.Amount,
		TransferProofURL:      req.TransferProofURL,
		TransferProofFilename: req.TransferProofFilename,
		SenderName:            req.SenderName,
		SenderAccountNumber:   req.SenderAccountNumber,
		TransferDate:          req.TransferDate,
		Status:                appPayment.ManualTransferStatusPending,
		DaoEntity: pubEntity.DaoEntity{
			Deleted:   false,
			CreatedAt: time.Now(),
		},
	}

	if err := dbTrx.GetManualTransferDAO().Insert(ctx, appPayment.ManualTransfers{transfer}); err != nil {
		return nil, fmt.Errorf("failed to insert transfer: %w", err)
	}

	if err := dbTrx.GetSqlTx().Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	return &transfer, nil
}

func (s *manualTransferService) GetPendingTransfers(ctx context.Context) ([]ManualTransferWithDetails, error) {
	dbTrx := dao.NewTransactionPayment(ctx, s.log, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	transfers, err := dbTrx.GetManualTransferDAO().Search(ctx, appPayment.ManualTransferQuery{
		Statuses: []string{string(appPayment.ManualTransferStatusPending)},
	})
	if err != nil {
		return nil, err
	}

	var result []ManualTransferWithDetails
	for _, t := range transfers {
		details, err := s.enrichTransferDetails(ctx, &t)
		if err != nil {
			s.log.Error(ctx, "Failed to enrich transfer details", zap.Error(err))
			continue
		}
		result = append(result, *details)
	}

	return result, nil
}

func (s *manualTransferService) GetTransferByOrderID(ctx context.Context, orderID string) (*appPayment.ManualTransfer, error) {
	dbTrx := dao.NewTransactionPayment(ctx, s.log, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()
	return dbTrx.GetManualTransferDAO().GetByOrderID(ctx, pubEntity.UUID(orderID))
}

func (s *manualTransferService) ApproveTransfer(ctx context.Context, transferID string, adminID string, notes string) error {
	dbTrx := dao.NewTransactionPayment(ctx, s.log, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	transfer, err := dbTrx.GetManualTransferDAO().GetByID(ctx, pubEntity.UUID(transferID))
	if err != nil {
		return err
	}
	if transfer == nil {
		return ErrManualTransferNotFound
	}
	if transfer.Status != appPayment.ManualTransferStatusPending {
		return ErrInvalidStatus
	}

	orders, err := dbTrx.GetOrderDAO().Search(ctx, orderEntity.OrderQuery{
		IDs: []string{string(transfer.OrderID)},
	})
	if err != nil || len(orders) == 0 {
		return ErrOrderNotFound
	}
	order := orders[0]

	registrants, _, err := dbTrx.GetRegistrantDAO().Search(ctx, regEntity.RegistrantQuery{
		IDs: []string{string(order.RegistrantID)},
	})
	if err != nil || len(registrants) == 0 {
		return errors.New("registrant not found")
	}
	registrant := registrants[0]

	attendees, err := dbTrx.GetAttendeeDAO().Search(ctx, regEntity.AttendeeQuery{
		RegistrantIDs: []string{string(registrant.ID)},
	})
	if err != nil {
		return err
	}

	ticketQtyMap := make(map[string]int)
	if registrant.TicketID != nil {
		ticketQtyMap[string(*registrant.TicketID)]++
	}
	for _, att := range attendees {
		ticketQtyMap[string(att.TicketID)]++
	}

	now := time.Now()
	for tID, qty := range ticketQtyMap {
		if err := dbTrx.GetTicketDAO().ConfirmSold(ctx, pubEntity.UUID(tID), qty); err != nil {
			return fmt.Errorf("failed to confirm sold: %w", err)
		}
	}

	order.PaymentStatus = "paid"
	order.PaymentTime = &now
	order.PaymentMethod = strPtr("MANUAL_TRANSFER")

	if err := dbTrx.GetOrderDAO().Update(ctx, []orderEntity.Order{order}); err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	registrant.Status = "paid"
	if err := dbTrx.GetRegistrantDAO().Update(ctx, []regEntity.Registrant{registrant}); err != nil {
		return fmt.Errorf("failed to update registrant: %w", err)
	}

	reviewedBy := pubEntity.UUID(adminID)
	transfer.Status = appPayment.ManualTransferStatusApproved
	transfer.ReviewedBy = &reviewedBy
	transfer.ReviewedAt = &now
	transfer.AdminNotes = &notes

	if err := dbTrx.GetManualTransferDAO().Update(ctx, appPayment.ManualTransfers{*transfer}); err != nil {
		return fmt.Errorf("failed to update transfer: %w", err)
	}

	if err := dbTrx.GetSqlTx().Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	s.sendTicketEmailAsync(ctx, order, registrant, ticketQtyMap)

	return nil
}

func (s *manualTransferService) RejectTransfer(ctx context.Context, transferID string, adminID string, notes string) error {
	dbTrx := dao.NewTransactionPayment(ctx, s.log, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	transfer, err := dbTrx.GetManualTransferDAO().GetByID(ctx, pubEntity.UUID(transferID))
	if err != nil {
		return err
	}
	if transfer == nil {
		return ErrManualTransferNotFound
	}
	if transfer.Status != appPayment.ManualTransferStatusPending {
		return ErrInvalidStatus
	}

	orders, err := dbTrx.GetOrderDAO().Search(ctx, orderEntity.OrderQuery{
		IDs: []string{string(transfer.OrderID)},
	})
	if err != nil || len(orders) == 0 {
		return ErrOrderNotFound
	}
	order := orders[0]

	registrants, _, err := dbTrx.GetRegistrantDAO().Search(ctx, regEntity.RegistrantQuery{
		IDs: []string{string(order.RegistrantID)},
	})
	if err != nil || len(registrants) == 0 {
		return errors.New("registrant not found")
	}
	registrant := registrants[0]

	now := time.Now()
	order.PaymentStatus = orderEntity.OrderStatusRejected
	order.PaymentMethod = strPtr("MANUAL_TRANSFER")

	if err := dbTrx.GetOrderDAO().Update(ctx, []orderEntity.Order{order}); err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	registrant.Status = orderEntity.OrderStatusRejected
	if err := dbTrx.GetRegistrantDAO().Update(ctx, []regEntity.Registrant{registrant}); err != nil {
		return fmt.Errorf("failed to update registrant: %w", err)
	}

	reviewedBy := pubEntity.UUID(adminID)
	transfer.Status = appPayment.ManualTransferStatusRejected
	transfer.ReviewedBy = &reviewedBy
	transfer.ReviewedAt = &now
	transfer.AdminNotes = &notes

	if err := dbTrx.GetManualTransferDAO().Update(ctx, appPayment.ManualTransfers{*transfer}); err != nil {
		return fmt.Errorf("failed to update transfer: %w", err)
	}

	if err := dbTrx.GetSqlTx().Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	s.sendRejectionEmailAsync(ctx, order, registrant, notes)

	return nil
}

func (s *manualTransferService) CancelTransfer(ctx context.Context, transferID string, adminID string, notes string) error {
	dbTrx := dao.NewTransactionPayment(ctx, s.log, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	transfer, err := dbTrx.GetManualTransferDAO().GetByID(ctx, pubEntity.UUID(transferID))
	if err != nil {
		return err
	}
	if transfer == nil {
		return ErrManualTransferNotFound
	}
	if transfer.Status != appPayment.ManualTransferStatusPending && transfer.Status != appPayment.ManualTransferStatusRejected {
		return ErrInvalidStatus
	}

	orders, err := dbTrx.GetOrderDAO().Search(ctx, orderEntity.OrderQuery{
		IDs: []string{string(transfer.OrderID)},
	})
	if err != nil || len(orders) == 0 {
		return ErrOrderNotFound
	}
	order := orders[0]

	registrants, _, err := dbTrx.GetRegistrantDAO().Search(ctx, regEntity.RegistrantQuery{
		IDs: []string{string(order.RegistrantID)},
	})
	if err != nil || len(registrants) == 0 {
		return errors.New("registrant not found")
	}
	registrant := registrants[0]

	attendees, err := dbTrx.GetAttendeeDAO().Search(ctx, regEntity.AttendeeQuery{
		RegistrantIDs: []string{string(registrant.ID)},
	})
	if err != nil {
		return err
	}

	ticketQtyMap := make(map[string]int)
	if registrant.TicketID != nil {
		ticketQtyMap[string(*registrant.TicketID)]++
	}
	for _, att := range attendees {
		ticketQtyMap[string(att.TicketID)]++
	}

	for tID, qty := range ticketQtyMap {
		if err := dbTrx.GetTicketDAO().ReleaseBooked(ctx, pubEntity.UUID(tID), qty); err != nil {
			s.log.Error(ctx, "failed to release booked tickets during cancellation", zap.String("ticket_id", tID), zap.Int("qty", qty), zap.Error(err))
		}
	}

	now := time.Now()
	order.PaymentStatus = orderEntity.OrderStatusFailed
	order.PaymentMethod = strPtr("MANUAL_TRANSFER")

	if err := dbTrx.GetOrderDAO().Update(ctx, []orderEntity.Order{order}); err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	registrant.Status = orderEntity.OrderStatusFailed
	if err := dbTrx.GetRegistrantDAO().Update(ctx, []regEntity.Registrant{registrant}); err != nil {
		return fmt.Errorf("failed to update registrant: %w", err)
	}

	reviewedBy := pubEntity.UUID(adminID)
	transfer.Status = appPayment.ManualTransferStatusCancelled
	transfer.ReviewedBy = &reviewedBy
	transfer.ReviewedAt = &now
	transfer.AdminNotes = &notes

	if err := dbTrx.GetManualTransferDAO().Update(ctx, appPayment.ManualTransfers{*transfer}); err != nil {
		return fmt.Errorf("failed to update transfer: %w", err)
	}

	if err := dbTrx.GetSqlTx().Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	s.sendCancellationEmailAsync(ctx, order, registrant, notes)

	return nil
}

func (s *manualTransferService) sendTicketEmailAsync(ctx context.Context, order orderEntity.Order, registrant regEntity.Registrant, ticketQtyMap map[string]int) {
	var ticketIDs []string
	for tID := range ticketQtyMap {
		ticketIDs = append(ticketIDs, tID)
	}

	ticketMap := make(map[string]ticketEntity.Ticket)
	if len(ticketIDs) > 0 {
		dbTrx := dao.NewTransactionPayment(ctx, s.log, s.sqlDB)
		tickets, err := dbTrx.GetTicketDAO().Search(ctx, ticketEntity.TicketQuery{
			IDs: ticketIDs,
		})
		if err != nil {
			s.log.Error(ctx, "Failed to fetch ticket data for PDF", zap.Error(err))
		} else {
			for _, t := range tickets {
				ticketMap[string(t.ID)] = t
			}
		}
	}

	dynamicEvent := orderSvc.EventDynamicData{
		EventName:      "Rakit Tiket Event",
		EventDate:      "Belum Ditentukan",
		EventTimeStart: "-",
		EventTimeEnd:   "-",
		EventLocation:  "Venue Terpilih",
	}

	_ = s.sqlDB.QueryRowContext(ctx, "SELECT name FROM events WHERE id = $1", order.EventID).Scan(&dynamicEvent.EventName)
	_ = s.sqlDB.QueryRowContext(ctx, "SELECT event_date, event_time_start, event_time_end, event_location FROM landing_pages WHERE event_id = $1", order.EventID).
		Scan(&dynamicEvent.EventDate, &dynamicEvent.EventTimeStart, &dynamicEvent.EventTimeEnd, &dynamicEvent.EventLocation)

	attachments, err := orderSvc.GenerateTicketsPDF(order, registrant, ticketMap, dynamicEvent)
	if err != nil {
		s.log.Error(ctx, "Failed to generate PDF tickets", zap.Error(err))
		return
	}

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

		err := s.emailService.SendTransferApprovalEmail(bgCtx, targetEmail, ordNum, evtName, ownerName, atts)
		if err != nil {
			s.log.Error(bgCtx, "Gagal mengirim email approval transfer", zap.Error(err))
		} else {
			s.log.Info(bgCtx, "Email approval transfer berhasil terkirim!", zap.String("to", targetEmail))
		}
	}(registrant.Email, order.OrderNumber, dynamicEvent.EventName, registrant.Name, emailAtts)
}

func (s *manualTransferService) sendRejectionEmailAsync(ctx context.Context, order orderEntity.Order, registrant regEntity.Registrant, reason string) {
	eventName := "Rakit Tiket Event"

	_ = s.sqlDB.QueryRowContext(ctx, "SELECT name FROM events WHERE id = $1", order.EventID).Scan(&eventName)

	go func(ordNum, evtName, ownerName, rejectionReason string) {
		bgCtx := context.Background()

		err := s.emailService.SendPaymentRejectedEmail(bgCtx, registrant.Email, ordNum, evtName, ownerName, rejectionReason)
		if err != nil {
			s.log.Error(bgCtx, "Gagal mengirim email penolakan transfer", zap.Error(err))
		} else {
			s.log.Info(bgCtx, "Email penolakan transfer berhasil terkirim!", zap.String("to", registrant.Email))
		}
	}(order.OrderNumber, eventName, registrant.Name, reason)
}

func (s *manualTransferService) sendCancellationEmailAsync(ctx context.Context, order orderEntity.Order, registrant regEntity.Registrant, reason string) {
	eventName := "Rakit Tiket Event"

	_ = s.sqlDB.QueryRowContext(ctx, "SELECT name FROM events WHERE id = $1", order.EventID).Scan(&eventName)

	go func(ordNum, evtName, ownerName, cancelReason string) {
		bgCtx := context.Background()

		err := s.emailService.SendPaymentCancelledEmail(bgCtx, registrant.Email, ordNum, evtName, ownerName, cancelReason)
		if err != nil {
			s.log.Error(bgCtx, "Gagal mengirim email pembatalan transfer", zap.Error(err))
		} else {
			s.log.Info(bgCtx, "Email pembatalan transfer berhasil terkirim!", zap.String("to", registrant.Email))
		}
	}(order.OrderNumber, eventName, registrant.Name, reason)
}

func (s *manualTransferService) enrichTransferDetails(ctx context.Context, transfer *appPayment.ManualTransfer) (*ManualTransferWithDetails, error) {
	dbTrx := dao.NewTransactionPayment(ctx, s.log, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	bankAccount, err := dbTrx.GetBankAccountDAO().GetByID(ctx, transfer.BankAccountID)
	if err != nil {
		return nil, err
	}

	orders, err := dbTrx.GetOrderDAO().Search(ctx, orderEntity.OrderQuery{
		IDs: []string{string(transfer.OrderID)},
	})
	if err != nil || len(orders) == 0 {
		return nil, ErrOrderNotFound
	}
	order := orders[0]

	registrants, _, err := dbTrx.GetRegistrantDAO().Search(ctx, regEntity.RegistrantQuery{
		IDs: []string{string(order.RegistrantID)},
	})
	if err != nil || len(registrants) == 0 {
		return nil, errors.New("registrant not found")
	}
	registrant := registrants[0]

	return &ManualTransferWithDetails{
		ManualTransfer: *transfer,
		BankAccount:    bankAccount,
		Order:          &order,
		Registrant:     &registrant,
	}, nil
}

type ManualTransferWithDetails struct {
	appPayment.ManualTransfer
	BankAccount *appPayment.BankAccount `json:"bank_account,omitempty"`
	Order       *orderEntity.Order      `json:"order,omitempty"`
	Registrant  interface{}             `json:"registrant,omitempty"`
}

type SubmitTransferProofRequest struct {
	OrderID               pubEntity.UUID `json:"order_id" validate:"required"`
	BankAccountID         pubEntity.UUID `json:"bank_account_id" validate:"required"`
	TransferProofURL      string         `json:"transfer_proof_url" validate:"required"`
	TransferProofFilename *string        `json:"transfer_proof_filename"`
	SenderName            string         `json:"sender_name" validate:"required"`
	SenderAccountNumber   *string        `json:"sender_account_number"`
	TransferDate          time.Time      `json:"transfer_date" validate:"required"`
}

func strPtr(s string) *string {
	return &s
}
