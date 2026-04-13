package model

import "time"

type CheckoutRequest struct {
	PaymentType string `json:"payment_type" validate:"required"`
}

type CheckoutResponse struct {
	OrderID       string            `json:"order_id"`
	OrderNumber   string            `json:"order_number"`
	Amount        float64           `json:"amount"`
	PaymentType   string            `json:"payment_type"`
	PaymentStatus string            `json:"payment_status"`
	ExpiresAt     *time.Time        `json:"expires_at"`
	PaymentInfo   *PaymentInfo      `json:"payment_info,omitempty"`
	BankAccounts  []BankAccountInfo `json:"bank_accounts,omitempty"`
}

type PaymentInfo struct {
	PaymentURL    string `json:"payment_url"`
	PaymentToken  string `json:"payment_token"`
	PaymentMethod string `json:"payment_method"`
}

type BankAccountInfo struct {
	BankName        string  `json:"bank_name"`
	BankCode        string  `json:"bank_code"`
	AccountNumber   string  `json:"account_number"`
	AccountHolder   string  `json:"account_holder"`
	InstructionText *string `json:"instruction_text"`
}
