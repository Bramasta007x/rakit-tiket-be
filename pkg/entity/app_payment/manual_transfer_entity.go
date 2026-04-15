package app_payment

import (
	"time"

	pubEntity "rakit-tiket-be/pkg/entity"
)

type ManualTransferStatus string

const (
	ManualTransferStatusPending   ManualTransferStatus = "PENDING"
	ManualTransferStatusApproved  ManualTransferStatus = "APPROVED"
	ManualTransferStatusRejected  ManualTransferStatus = "REJECTED"
	ManualTransferStatusCancelled ManualTransferStatus = "CANCELLED"
)

type ManualTransfer struct {
	ID                    pubEntity.UUID       `json:"id"`
	OrderID               pubEntity.UUID       `json:"order_id"`
	BankAccountID         pubEntity.UUID       `json:"bank_account_id"`
	TransferAmount        float64              `json:"transfer_amount"`
	TransferProofURL      string               `json:"transfer_proof_url"`
	TransferProofFilename *string              `json:"transfer_proof_filename"`
	SenderName            string               `json:"sender_name"`
	SenderAccountNumber   *string              `json:"sender_account_number"`
	TransferDate          time.Time            `json:"transfer_date"`
	AdminNotes            *string              `json:"admin_notes"`
	ReviewedBy            *pubEntity.UUID      `json:"reviewed_by"`
	ReviewedAt            *time.Time           `json:"reviewed_at"`
	Status                ManualTransferStatus `json:"status"`
	pubEntity.DaoEntity

	BankAccount *BankAccount `json:"bank_account,omitempty"`
}

type ManualTransfers []ManualTransfer

type ManualTransferQuery struct {
	IDs      []string `query:"id"`
	OrderIDs []string `query:"order_id"`
	Statuses []string `query:"status"`
	EventIDs []string `query:"event_id"`
}
