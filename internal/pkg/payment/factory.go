package payment

import (
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
	// case GatewayXendit:
	//	return xendit_gateway.NewXenditProvider(...), nil
	default:
		return nil, fmt.Errorf("unsupported payment gateway: %s", gateway)
	}
}
