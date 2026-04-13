package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"rakit-tiket-be/internal/app/app_payment/dao"
	"rakit-tiket-be/internal/app/app_payment/model"
	"rakit-tiket-be/internal/pkg/payment"
	"rakit-tiket-be/pkg/entity/app_order"
	appPayment "rakit-tiket-be/pkg/entity/app_payment"
	appRegistrant "rakit-tiket-be/pkg/entity/app_registrant"
	ticketEntity "rakit-tiket-be/pkg/entity/app_ticket"
	"rakit-tiket-be/pkg/util"
)

type CheckoutInitiator interface {
	InitiateGatewayPayment(ctx context.Context, order *app_order.Order) (*InitiateResult, error)
	InitiateManualTransfer(ctx context.Context, order *app_order.Order) (*InitiateResult, error)
}

type InitiateResult struct {
	PaymentType  string
	GatewayCode  *string
	PaymentInfo  *model.PaymentInfo
	BankAccounts []model.BankAccountInfo
}

type checkoutInitiator struct {
	log            util.LogUtil
	sqlDB          *sql.DB
	paymentFactory *payment.PaymentFactory
	bankAccountSvc BankAccountServiceProvider
}

func MakeCheckoutInitiator(log util.LogUtil, sqlDB *sql.DB, paymentFactory *payment.PaymentFactory, bankAccountSvc BankAccountServiceProvider) CheckoutInitiator {
	return &checkoutInitiator{
		log:            log,
		sqlDB:          sqlDB,
		paymentFactory: paymentFactory,
		bankAccountSvc: bankAccountSvc,
	}
}

func (s *checkoutInitiator) InitiateGatewayPayment(ctx context.Context, order *app_order.Order) (*InitiateResult, error) {
	dbTrx := dao.NewTransactionPayment(ctx, s.log, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	gateways, err := dbTrx.GetGatewayDAO().Search(ctx, appPayment.GatewayQuery{
		IsActive:  func() *bool { v := true; return &v }(),
		IsEnabled: func() *bool { v := true; return &v }(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get active gateway: %v", err)
	}

	if len(gateways) == 0 {
		return nil, errors.New("no active payment gateway configured")
	}

	activeGateway := gateways[0]

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

	provider, err := s.paymentFactory.GetProviderByCode(activeGateway.Code)
	if err != nil {
		return nil, fmt.Errorf("unsupported gateway: %s", activeGateway.Code)
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

	gatewayCode := activeGateway.Code
	order.PaymentGateway = &gatewayCode
	order.PaymentType = func() *string { v := app_order.PaymentTypeGateway; return &v }()
	order.PaymentToken = &paymentResp.Token
	order.PaymentURL = &paymentResp.RedirectURL

	if err := dbTrx.GetOrderDAO().Update(ctx, []app_order.Order{*order}); err != nil {
		return nil, err
	}

	if err := dbTrx.GetSqlTx().Commit(); err != nil {
		return nil, err
	}

	return &InitiateResult{
		PaymentType: app_order.PaymentTypeGateway,
		GatewayCode: &gatewayCode,
		PaymentInfo: &model.PaymentInfo{
			PaymentURL:   paymentResp.RedirectURL,
			PaymentToken: paymentResp.Token,
		},
	}, nil
}

func (s *checkoutInitiator) InitiateManualTransfer(ctx context.Context, order *app_order.Order) (*InitiateResult, error) {
	dbTrx := dao.NewTransactionPayment(ctx, s.log, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	paymentTypeStr := app_order.PaymentTypeManual
	order.PaymentType = &paymentTypeStr

	if err := dbTrx.GetOrderDAO().Update(ctx, []app_order.Order{*order}); err != nil {
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

	return &InitiateResult{
		PaymentType:  app_order.PaymentTypeManual,
		BankAccounts: bankAccounts,
	}, nil
}

func IsGatewayPaymentType(paymentType string) bool {
	gatewayCodes := []string{"MIDTRANS", "XENDIT", "DOKU", "GATEWAY"}
	for _, code := range gatewayCodes {
		if strings.ToUpper(paymentType) == code {
			return true
		}
	}
	return false
}
