package payment

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
	"github.com/midtrans/midtrans-go/snap"
)

type midtransProvider struct {
	snapClient snap.Client
	coreClient coreapi.Client
}

func NewMidtransProvider(serverKey string, isProduction bool) Provider {
	var env midtrans.EnvironmentType = midtrans.Sandbox
	if isProduction {
		env = midtrans.Production
	}

	var s snap.Client
	s.New(serverKey, env)

	var c coreapi.Client
	c.New(serverKey, env)

	return &midtransProvider{
		snapClient: s,
		coreClient: c,
	}
}

// Implementasi fungsi CreateTransaction dari interface Provider
func (m *midtransProvider) CreateTransaction(ctx context.Context, req CreateTransactionRequest) (*CreateTransactionResponse, error) {

	// Mapping dari Format Universal ke Format Midtrans
	midtransItems := []midtrans.ItemDetails{}
	for _, item := range req.Items {
		midtransItems = append(midtransItems, midtrans.ItemDetails{
			ID:    item.ID,
			Name:  item.Name,
			Price: int64(item.Price),
			Qty:   int32(item.Quantity),
		})
	}

	snapReq := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  req.OrderID,
			GrossAmt: int64(req.Amount),
		},
		CustomerDetail: &midtrans.CustomerDetails{
			FName: req.Customer.Name,
			Email: req.Customer.Email,
			Phone: req.Customer.Phone,
		},
		Items: &midtransItems,
		Expiry: &snap.ExpiryDetails{
			Unit:     "minute",
			Duration: int64(req.ExpiryMinutes),
		},
	}

	// Eksekusi ke midtrans
	snapResp, err := m.snapClient.CreateTransaction(snapReq)
	if err != nil {
		return nil, fmt.Errorf("midtrans error: %v", err.GetMessage())
	}

	// Mapping balik ke format universal
	return &CreateTransactionResponse{
		Token:       snapResp.Token,
		RedirectURL: snapResp.RedirectURL,
	}, nil
}

// Implementasi fungsi Webhook
func (m *midtransProvider) ParseWebhook(ctx context.Context, payload []byte) (*WebhookNotification, error) {
	var notif map[string]interface{}
	if err := json.Unmarshal(payload, notif); err != nil {
		return nil, err
	}

	// Mapping logic webhook midtrans
	orderID, _ := notif["order_id"].(string)
	transactionID, _ := notif["transaction_id"].(string)
	transactionStatus, _ := notif["transaction_status"].(string)
	paymentType, _ := notif["payment_type"].(string)

	// Ubah status midtrans ke status universal
	var mappedStatus string
	switch transactionStatus {
	case "settlement", "capture":
		mappedStatus = "paid"
	case "pending":
		mappedStatus = "pending"
	case "deny", "cancel", "expire", "failure":
		mappedStatus = "failed"
	default:
		mappedStatus = transactionStatus
	}

	rawPayload, _ := json.Marshal(notif)

	return &WebhookNotification{
		OrderID:       orderID,
		TransactionID: transactionID,
		PaymentStatus: mappedStatus,
		PaymentType:   paymentType,
		Gateway:       GatewayMidtrans,
		RawPayload:    string(rawPayload),
	}, nil
}
