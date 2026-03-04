package app_registrant

import (
	"time"

	pubEntity "rakit-tiket-be/pkg/entity"
)

type (
	AttendeeQuery struct {
		IDs           []string `query:"id"`
		EventIDs      []string `query:"event_id"`
		RegistrantIDs []string `query:"registrant_id"`
		TicketIDs     []string `query:"ticket_id"`
	}

	Attendee struct {
		ID      pubEntity.UUID `json:"id"`
		EventID pubEntity.UUID `json:"event_id"`

		RegistrantID pubEntity.UUID `json:"registrant_id"`
		TicketID     pubEntity.UUID `json:"ticket_id"`

		// Personal Info
		Name      string     `json:"name"`
		Gender    *string    `json:"gender"`
		Birthdate *time.Time `json:"birthdate"`

		pubEntity.DaoEntity
	}

	Attendees []Attendee
)
