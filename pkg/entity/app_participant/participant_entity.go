package app_participant

import (
	"time"

	pubEntity "rakit-tiket-be/pkg/entity"
)

type (
	ParticipantFilter struct {
		EventID       string   `query:"event_id"`
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

	ParticipantPayment struct {
		ID            pubEntity.UUID `json:"id"`
		Status        string         `json:"status"`
		Method        string         `json:"method"`
		PaymentMethod string         `json:"payment_method"`
		Time          *time.Time     `json:"time"`
		Total         float64        `json:"total"`
	}

	ParticipantRegistrant struct {
		Name        string     `json:"name"`
		TicketTitle string     `json:"ticket_title"`
		TicketType  string     `json:"ticket_type"`
		Email       string     `json:"email"`
		Phone       string     `json:"phone"`
		Gender      string     `json:"gender"`
		Birthdate   *time.Time `json:"birthdate"`
		ETicket     *string    `json:"e_ticket"`
	}

	ParticipantAttendee struct {
		TicketTitle string     `json:"ticket_title"`
		TicketType  string     `json:"ticket_type"`
		Name        string     `json:"name"`
		Gender      string     `json:"gender"`
		Birthdate   *time.Time `json:"birthdate"`
		ETicket     *string    `json:"e_ticket"`
	}

	Participant struct {
		UniqueID     pubEntity.UUID        `json:"unique_id"`
		Payment      ParticipantPayment    `json:"payment"`
		OrderNumber  string                `json:"order_number"`
		Registrant   ParticipantRegistrant `json:"registrant"`
		Attendees    []ParticipantAttendee `json:"attendees"`
		TotalTickets int                   `json:"total_tickets"`
		TicketTypes  []string              `json:"ticket_types"`
		TicketTitles []string              `json:"ticket_titles"`
	}

	Participants []Participant

	ParticipantListResponse struct {
		Data       Participants `json:"data"`
		TotalCount int          `json:"totalCount"`
		Page       int          `json:"page"`
		PerPage    int          `json:"per_page"`
		SortBy     string       `json:"sort_by"`
		SortOrder  string       `json:"sort_order"`
	}
)

type (
	ParticipantDAOData struct {
		RegistrantID          pubEntity.UUID
		RegistrantName        string
		RegistrantEmail       string
		RegistrantPhone       string
		RegistrantGender      *string
		RegistrantBirthdate   *time.Time
		RegistrantTicketID    *pubEntity.UUID
		RegistrantTicketTitle *string
		RegistrantTicketType  *string

		OrderID             pubEntity.UUID
		OrderNumber         *string
		OrderAmount         *float64
		OrderPaymentStatus  *string
		OrderPaymentMethod  *string
		OrderPaymentChannel *string
		OrderPaymentTime    *time.Time

		AttendeeID          pubEntity.UUID
		AttendeeName        string
		AttendeeGender      *string
		AttendeeBirthdate   *time.Time
		AttendeeTicketID    pubEntity.UUID
		AttendeeTicketTitle *string
		AttendeeTicketType  *string

		CreatedAt time.Time
	}
)
