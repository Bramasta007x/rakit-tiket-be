package app_registrant

import (
	"time"

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

type ListFilter struct {
	Search        string   `query:"search"`
	TicketType    []string `query:"ticket_type"`
	PaymentStatus []string `query:"payment_status"`
	DateStart     string   `query:"date_start"`
	DateEnd       string   `query:"date_end"`
	SortBy        string   `query:"sort_by"`
	SortOrder     string   `query:"sort_order"`
	PerPage       int      `query:"per_page"`
	Page          int      `query:"page"`
}

type PaymentInfo struct {
	ID            string     `json:"id"`
	Status        string     `json:"status"`
	Method        string     `json:"method"`
	PaymentMethod string     `json:"payment_method"`
	Time          *time.Time `json:"time"`
	Total         float64    `json:"total"`
}

type RegistrantInfoDetail struct {
	Name        string  `json:"name"`
	TicketTitle string  `json:"ticket_title"`
	TicketType  string  `json:"ticket_type"`
	Email       string  `json:"email"`
	Phone       string  `json:"phone"`
	Gender      string  `json:"gender"`
	Birthdate   *string `json:"birthdate"`
	ETicket     *string `json:"e_ticket"`
}

type AttendeeInfo struct {
	TicketTitle string  `json:"ticket_title"`
	TicketType  string  `json:"ticket_type"`
	Name        string  `json:"name"`
	Gender      string  `json:"gender"`
	Birthdate   *string `json:"birthdate"`
	ETicket     *string `json:"e_ticket"`
}

type ListItem struct {
	UniqueID     string               `json:"unique_id"`
	Payment      PaymentInfo          `json:"payment"`
	OrderNumber  string               `json:"order_number"`
	Registrant   RegistrantInfoDetail `json:"registrant"`
	Attendees    []AttendeeInfo       `json:"attendees"`
	TotalTickets int                  `json:"total_tickets"`
	TicketTypes  []string             `json:"ticket_types"`
	TicketTitles []string             `json:"ticket_titles"`
}

type ListResponse struct {
	Data       []ListItem `json:"data"`
	Total      int        `json:"total"`
	Page       int        `json:"page"`
	PerPage    int        `json:"per_page"`
	TotalPages int        `json:"total_pages"`
}
