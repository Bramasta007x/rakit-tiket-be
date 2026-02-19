package app_registrant

import (
	"rakit-tiket-be/pkg/entity"
)

type RegistrantData struct {
	TicketID  entity.UUID `json:"ticket_id" validate:"required"`
	Name      string      `json:"name" validate:"required"`
	Email     string      `json:"email" validate:"required,email"`
	Phone     string      `json:"phone" validate:"required"`
	Gender    *string     `json:"gender"`
	Birthdate *string     `json:"birthdate"` // Format: YYYY-MM-DD
	Document  *string     `json:"document"`
}

type AttendeeData struct {
	TicketID  entity.UUID `json:"ticket_id" validate:"required"`
	Name      string      `json:"name" validate:"required"`
	Gender    *string     `json:"gender"`
	Birthdate *string     `json:"birthdate"` // Format: YYYY-MM-DD
	Document  *string     `json:"document"`
}

type RegisterRequest struct {
	Registrant RegistrantData `json:"registrant" validate:"required"`
	Attendees  []AttendeeData `json:"attendees"`
}

type RegisterResponse struct {
	Order      OrderInfo      `json:"order"`
	Registrant RegistrantInfo `json:"registrant"`
}

type OrderInfo struct {
	OrderID       string  `json:"order_id"`
	OrderNumber   string  `json:"order_number"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	PaymentStatus string  `json:"payment_status"`
	PaymentToken  string  `json:"payment_token"`
	RedirectURL   string  `json:"redirect_url"`
}

type RegistrantInfo struct {
	ID         string `json:"id"`
	UniqueCode string `json:"unique_code"`
}
