package app_order

import (
	"time"

	pubEntity "rakit-tiket-be/pkg/entity"
)

// Payment Status Constants
const (
	OrderStatusPending = "pending"
	OrderStatusPaid    = "paid"
	OrderStatusExpired = "expired"
	OrderStatusFailed  = "failed"
)

// Payment Type Constants
const (
	PaymentTypeGateway = "GATEWAY"
	PaymentTypeManual  = "MANUAL"
)

type (
	OrderQuery struct {
		IDs             []string   `query:"id"`
		EventIDs        []string   `query:"event_id"`
		OrderNumbers    []string   `query:"order_number"`
		RegistrantIDs   []string   `query:"registrant_id"`
		PaymentGateways []string   `query:"payment_gateway"`
		Statuses        []string   `query:"payment_status"`
		ExpiredBefore   *time.Time `query:"expired_before"`
	}

	Order struct {
		ID      pubEntity.UUID `json:"id"`
		EventID pubEntity.UUID `json:"event_id"`

		// Relations
		RegistrantID pubEntity.UUID `json:"registrant_id"`

		// Transaction Details
		OrderNumber string  `json:"order_number"`
		Amount      float64 `json:"amount"`
		Currency    string  `json:"currency"`

		// Payment Type: GATEWAY (Midtrans) atau MANUAL (Transfer)
		PaymentType *string `json:"payment_type"`

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

		// Payment Proof (untuk MANUAL transfer)
		PaymentProofURL      *string `json:"payment_proof_url"`
		PaymentProofFilename *string `json:"payment_proof_filename"`

		// Verification (Admin Audit Trail)
		VerifiedBy *pubEntity.UUID `json:"verified_by"`
		VerifiedAt *time.Time      `json:"verified_at"`

		PaymentTime *time.Time `json:"payment_time"`
		ExpiresAt   *time.Time `json:"expires_at"`

		// Metadata
		pubEntity.DaoEntity
	}

	Orders []Order
)
