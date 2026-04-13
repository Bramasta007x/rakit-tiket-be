package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"rakit-tiket-be/internal/app/app_payment/model"
	regDao "rakit-tiket-be/internal/app/app_registrant/dao"
	"rakit-tiket-be/internal/pkg/payment"
	"rakit-tiket-be/pkg/entity/app_order"
	appPayment "rakit-tiket-be/pkg/entity/app_payment"
	appRegistrant "rakit-tiket-be/pkg/entity/app_registrant"
	ticketEntity "rakit-tiket-be/pkg/entity/app_ticket"
	"rakit-tiket-be/pkg/util"
)

type CheckoutService interface {
	InitiateCheckout(ctx context.Context, orderID string, paymentType string) (*model.CheckoutResponse, error)
	GetActivePaymentOptions(ctx context.Context) ([]PaymentOption, error)
}

type checkoutService struct {
	log              util.LogUtil
	sqlDB            *sql.DB
	paymentFactory   *payment.PaymentFactory
	bankAccountSvc   BankAccountServiceProvider
	paymentConfigSvc PaymentConfigProvider
}

type BankAccountServiceProvider interface {
	GetActiveBankAccounts(ctx context.Context) (appPayment.BankAccounts, error)
}

type PaymentConfigProvider interface {
	GetActiveGateway(ctx context.Context) (*appPayment.PaymentGateway, error)
	GetActivePaymentOptions(ctx context.Context) ([]PaymentOption, error)
	IsManualTransferEnabled(ctx context.Context) (bool, error)
}

func MakeCheckoutService(log util.LogUtil, sqlDB *sql.DB, paymentFactory *payment.PaymentFactory, bankAccountSvc BankAccountServiceProvider, paymentConfigSvc PaymentConfigProvider) CheckoutService {
	return &checkoutService{
		log:              log,
		sqlDB:            sqlDB,
		paymentFactory:   paymentFactory,
		bankAccountSvc:   bankAccountSvc,
		paymentConfigSvc: paymentConfigSvc,
	}
}

func (s *checkoutService) InitiateCheckout(ctx context.Context, orderID string, paymentType string) (*model.CheckoutResponse, error) {
	dbTrx := regDao.NewTransactionRegistrant(ctx, s.log, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	orders, err := dbTrx.GetOrderDAO().Search(ctx, app_order.OrderQuery{
		IDs: []string{orderID},
	})
	if err != nil || len(orders) == 0 {
		return nil, errors.New("order not found")
	}
	order := orders[0]

	if order.PaymentStatus != app_order.OrderStatusPending {
		return nil, fmt.Errorf("order cannot be checked out (status: %s)", order.PaymentStatus)
	}

	now := time.Now()
	if order.ExpiresAt != nil && now.After(*order.ExpiresAt) {
		return nil, errors.New("order has expired")
	}

	paymentTypeUpper := strings.ToUpper(paymentType)

	if s.isGatewayPaymentType(paymentTypeUpper) {
		gateway, err := s.paymentConfigSvc.GetActiveGateway(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get active gateway: %v", err)
		}
		if gateway == nil {
			return nil, errors.New("no active payment gateway configured")
		}

		registrants, _, err := dbTrx.GetRegistrantDAO().Search(ctx, appRegistrant.RegistrantQuery{
			IDs: []string{string(order.RegistrantID)},
		})
		if err != nil || len(registrants) == 0 {
			return nil, errors.New("registrant not found")
		}
		reg := registrants[0]

		attendees, err := dbTrx.GetAttendeeDAO().Search(ctx, appRegistrant.AttendeeQuery{
			RegistrantIDs: []string{string(order.RegistrantID)},
		})
		if err != nil {
			return nil, err
		}

		ticketIDs := []string{}
		if reg.TicketID != nil {
			ticketIDs = append(ticketIDs, string(*reg.TicketID))
		}
		for _, att := range attendees {
			ticketIDs = append(ticketIDs, string(att.TicketID))
		}

		ticketQtyMap := make(map[string]int)
		for _, tID := range ticketIDs {
			ticketQtyMap[tID]++
		}

		var uniqueTicketIDs []string
		for tID := range ticketQtyMap {
			uniqueTicketIDs = append(uniqueTicketIDs, tID)
		}

		tickets, err := dbTrx.GetTicketDAO().Search(ctx, ticketEntity.TicketQuery{
			IDs: uniqueTicketIDs,
		})
		if err != nil {
			return nil, err
		}

		ticketMap := make(map[string]ticketEntity.Ticket)
		for _, t := range tickets {
			ticketMap[string(t.ID)] = t
		}

		var paymentItems []payment.Item
		for tID, qty := range ticketQtyMap {
			if t, ok := ticketMap[tID]; ok {
				paymentItems = append(paymentItems, payment.Item{
					ID:       string(t.ID),
					Name:     t.Title,
					Price:    t.Price,
					Quantity: qty,
				})
			}
		}

		provider, err := s.paymentFactory.GetProviderByCode(gateway.Code)
		if err != nil {
			return nil, fmt.Errorf("unsupported gateway: %s", gateway.Code)
		}

		var expiryMinutes int = 15
		if order.ExpiresAt != nil {
			expiryMinutes = int(time.Until(*order.ExpiresAt).Minutes())
			if expiryMinutes < 1 {
				expiryMinutes = 1
			}
		}

		paymentReq := payment.CreateTransactionRequest{
			OrderID:       order.OrderNumber,
			Amount:        order.Amount,
			Customer:      payment.Customer{Name: reg.Name, Email: reg.Email, Phone: reg.Phone},
			Items:         paymentItems,
			ExpiryMinutes: expiryMinutes,
		}

		paymentResp, err := provider.CreateTransaction(ctx, paymentReq)
		if err != nil {
			return nil, fmt.Errorf("payment gateway error: %v", err)
		}

		gatewayCode := gateway.Code
		paymentTypeStr := app_order.PaymentTypeGateway

		order.PaymentGateway = &gatewayCode
		order.PaymentType = &paymentTypeStr
		order.PaymentToken = &paymentResp.Token
		order.PaymentURL = &paymentResp.RedirectURL

		if err := dbTrx.GetOrderDAO().Update(ctx, []app_order.Order{order}); err != nil {
			return nil, err
		}

		if err := dbTrx.GetSqlTx().Commit(); err != nil {
			return nil, err
		}

		return &model.CheckoutResponse{
			OrderID:       string(order.ID),
			OrderNumber:   order.OrderNumber,
			Amount:        order.Amount,
			PaymentType:   *order.PaymentType,
			PaymentStatus: order.PaymentStatus,
			ExpiresAt:     order.ExpiresAt,
			PaymentInfo: &model.PaymentInfo{
				PaymentURL:   paymentResp.RedirectURL,
				PaymentToken: paymentResp.Token,
			},
		}, nil
	}

	if paymentTypeUpper == "MANUAL" {
		manualEnabled, err := s.paymentConfigSvc.IsManualTransferEnabled(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to check manual transfer status: %v", err)
		}
		if !manualEnabled {
			return nil, errors.New("manual transfer is not enabled")
		}

		paymentTypeStr := app_order.PaymentTypeManual
		order.PaymentType = &paymentTypeStr

		if err := dbTrx.GetOrderDAO().Update(ctx, []app_order.Order{order}); err != nil {
			return nil, err
		}

		accounts, err := s.bankAccountSvc.GetActiveBankAccounts(ctx)
		if err != nil {
			return nil, err
		}

		var bankAccounts []model.BankAccountInfo
		for _, acc := range accounts {
			bankAccounts = append(bankAccounts, model.BankAccountInfo{
				BankName:        acc.BankName,
				BankCode:        acc.BankCode,
				AccountNumber:   acc.AccountNumber,
				AccountHolder:   acc.AccountHolder,
				InstructionText: acc.InstructionText,
			})
		}

		if err := dbTrx.GetSqlTx().Commit(); err != nil {
			return nil, err
		}

		return &model.CheckoutResponse{
			OrderID:       string(order.ID),
			OrderNumber:   order.OrderNumber,
			Amount:        order.Amount,
			PaymentType:   paymentTypeStr,
			PaymentStatus: order.PaymentStatus,
			ExpiresAt:     order.ExpiresAt,
			BankAccounts:  bankAccounts,
		}, nil
	}

	return nil, errors.New("invalid payment type")
}

func (s *checkoutService) isGatewayPaymentType(paymentType string) bool {
	gatewayCodes := []string{"MIDTRANS", "XENDIT", "DOKU", "GATEWAY"}
	for _, code := range gatewayCodes {
		if paymentType == code {
			return true
		}
	}
	return false
}

func (s *checkoutService) GetActivePaymentOptions(ctx context.Context) ([]PaymentOption, error) {
	return s.paymentConfigSvc.GetActivePaymentOptions(ctx)
}
