package app_registrant

import (
	"time"

	pubEntity "rakit-tiket-be/pkg/entity"
)

type (
	RegistrantQuery struct {
		IDs         []string `query:"id"`
		UniqueCodes []string `query:"unique_code"`
		Emails      []string `query:"email"`
		Statuses    []string `query:"status"`
	}

	Registrant struct {
		ID pubEntity.UUID `json:"id"`

		// Registrant Info
		UniqueCode string          `json:"unique_code"`
		TicketID   *pubEntity.UUID `json:"ticket_id"`
		Name       string          `json:"name"`
		Email      string          `json:"email"`
		Phone      string          `json:"phone"`
		Gender     *string         `json:"gender"`
		Birthdate  *time.Time      `json:"birthdate"`

		// Transaction Info
		TotalCost    float64 `json:"total_cost"`
		TotalTickets int     `json:"total_tickets"`
		Status       string  `json:"status"`

		Attendees Attendee `json:"attendee"`

		pubEntity.DaoEntity
	}

	Registrants []Registrant
)
