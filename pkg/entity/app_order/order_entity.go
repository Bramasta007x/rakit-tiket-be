package app_order

import (
	"time"

	pubEntity "rakit-tiket-be/pkg/entity"
)

type (
	OrderQuery struct {
		IDs             []string `query:"id"`
		OrderNumbers    []string `query:"order_number"`
		RegistrantIDs   []string `query:"registrant_id"`
		PaymentGateways []string `query:"payment_gateway"`
		Statuses        []string `query:"payment_status"`
	}

	Order struct {
		ID pubEntity.UUID `json:"id"`

		// Relations
		RegistrantID pubEntity.UUID `json:"registrant_id"`

		// Transaction Details
		OrderNumber string  `json:"order_number"`
		Amount      float64 `json:"amount"`
		Currency    string  `json:"currency"`

		// Generic Payment Info
		PaymentGateway *string `json:"payment_gateway"` // MIDTRANS, XENDIT, dll
		PaymentMethod  *string `json:"payment_method"`
		PaymentChannel *string `json:"payment_channel"`
		PaymentStatus  string  `json:"payment_status"`

		// Gateway Specific Data (Generic Names)
		PaymentToken         *string `json:"payment_token"`
		PaymentURL           *string `json:"payment_url"`
		PaymentTransactionID *string `json:"payment_transaction_id"`
		PaymentMetadata      *string `json:"payment_metadata"` // JSON string

		PaymentTime *time.Time `json:"payment_time"`
		ExpiresAt   *time.Time `json:"expires_at"`

		// Metadata
		pubEntity.DaoEntity
	}

	Orders []Order
)
