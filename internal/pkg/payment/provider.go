package payment

import "context"

type GatewayType string

const (
	GatewayMidtrans GatewayType = "MIDTRANS"
	GatewayXendit   GatewayType = "XENDIT"
)

// DTO Request (Universal)
type Customer struct {
	Name  string
	Email string
	Phone string
}

type Item struct {
	ID       string
	Name     string
	Price    float64
	Quantity int
}

type CreateTransactionRequest struct {
	OrderID  string
	Amount   float64
	Customer Customer
	Items    []Item
}

// DTO Response (Universal)
type CreateTransactionResponse struct {
	Token         string // Contoh: snap_token Midtrans, invoice_id Xendit
	RedirectURL   string // URL untuk dilempar ke frontend
	TransactionID string // ID transaksi dari sistem gateway (jika langsung digenerate)
}

// DTO Webhook (Universal)
type WebhookNotification struct {
	OrderID       string
	TransactionID string
	PaymentStatus string // "paid", "pending", "failed", "expired"
	PaymentType   string // "bank_transfer", "gopay", "credit_card"
	Gateway       GatewayType
	RawPayload    string // Disimpan ke `payment_metadata` untuk tracking
}

// 4. THE STRATEGY INTERFACE
type Provider interface {
	CreateTransaction(ctx context.Context, req CreateTransactionRequest) (*CreateTransactionResponse, error)
	ParseWebhook(ctx context.Context, payload []byte) (*WebhookNotification, error)
}
