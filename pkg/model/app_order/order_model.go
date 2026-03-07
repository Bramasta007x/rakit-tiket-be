package model

import "time"

type OrderStatusResponse struct {
	OrderNumber    string           `json:"order_number"`
	PaymentMethod  string           `json:"payment_method"`
	PaymentChannel string           `json:"payment_channel"`
	PaymentStatus  string           `json:"payment_status"`
	Amount         float64          `json:"amount"`
	PaymentTime    *time.Time       `json:"payment_time"`
	Registrant     RegistrantStatus `json:"registrant"`
	Attendees      []AttendeeStatus `json:"attendees"`
}

type RegistrantStatus struct {
	Name        string  `json:"name"`
	Email       string  `json:"email"`
	Phone       string  `json:"phone"`
	Gender      *string `json:"gender"`
	Birthdate   *string `json:"birthdate"`
	TicketTitle *string `json:"ticket_title"`
	TicketType  *string `json:"ticket_type"`
}

type AttendeeStatus struct {
	Name        string  `json:"name"`
	Gender      *string `json:"gender"`
	Birthdate   *string `json:"birthdate"`
	TicketTitle *string `json:"ticket_title"`
	TicketType  *string `json:"ticket_type"`
}
