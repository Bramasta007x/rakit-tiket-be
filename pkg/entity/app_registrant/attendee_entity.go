package app_registrant

import (
	"time"

	pubEntity "rakit-tiket-be/pkg/entity"
)

type (
	AttendeeQuery struct {
		IDs           []string `query:"id"`
		RegistrantIDs []string `query:"registrant_id"`
		TicketIDs     []string `query:"ticket_id"`
	}

	Attendee struct {
		ID pubEntity.UUID `json:"id"`

		RegistrantID pubEntity.UUID `json:"registrant_id"`
		TicketID     pubEntity.UUID `json:"ticket_id"`

		// Personal Info
		Name      string     `json:"name"`
		Gender    *string    `json:"gender"`
		Birthdate *time.Time `json:"birthdate"`

		// Metadata (Deleted, DataHash, CreatedAt, UpdatedAt)
		pubEntity.DaoEntity
	}

	Attendees []Attendee
)
