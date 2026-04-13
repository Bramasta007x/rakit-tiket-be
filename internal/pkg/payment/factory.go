package payment

import (
	"context"
	"fmt"
)

type PaymentFactory struct {
	midtransServerKey string
	isProduction      bool
	// Xendit, Dokus inisiasi disini
}

func NewPaymentFactory(midtransServerKey string, isProd bool) *PaymentFactory {
	return &PaymentFactory{
		midtransServerKey: midtransServerKey,
		isProduction:      isProd,
	}
}

func (f *PaymentFactory) GetProvider(gateway GatewayType) (Provider, error) {
	switch gateway {
	case GatewayMidtrans:
		return NewMidtransProvider(f.midtransServerKey, f.isProduction), nil
	case GatewayXendit:
		return NewXenditProvider(), nil
	case GatewayDoku:
		return NewDokuProvider(), nil
	default:
		return nil, fmt.Errorf("unsupported payment gateway: %s", gateway)
	}
}

func (f *PaymentFactory) GetProviderByCode(code string) (Provider, error) {
	return f.GetProvider(GatewayType(code))
}

func NewXenditProvider() Provider {
	return &xenditProvider{}
}

func NewDokuProvider() Provider {
	return &dokuProvider{}
}

type xenditProvider struct{}

func (p *xenditProvider) CreateTransaction(ctx context.Context, req CreateTransactionRequest) (*CreateTransactionResponse, error) {
	return nil, fmt.Errorf("xendit provider not implemented")
}

func (p *xenditProvider) ParseWebhook(ctx context.Context, payload []byte) (*WebhookNotification, error) {
	return nil, fmt.Errorf("xendit webhook not implemented")
}

type dokuProvider struct{}

func (p *dokuProvider) CreateTransaction(ctx context.Context, req CreateTransactionRequest) (*CreateTransactionResponse, error) {
	return nil, fmt.Errorf("doku provider not implemented")
}

func (p *dokuProvider) ParseWebhook(ctx context.Context, payload []byte) (*WebhookNotification, error) {
	return nil, fmt.Errorf("doku webhook not implemented")
}
