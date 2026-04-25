package app_physical_ticket

import (
	"time"

	pubEntity "rakit-tiket-be/pkg/entity"
)

type PhysicalTicketStatus string

const (
	PhysicalTicketStatusActive     PhysicalTicketStatus = "ACTIVE"
	PhysicalTicketStatusCheckedIn  PhysicalTicketStatus = "CHECKED_IN"
	PhysicalTicketStatusCheckedOut PhysicalTicketStatus = "CHECKED_OUT"
	PhysicalTicketStatusExceeded   PhysicalTicketStatus = "EXCEEDED"
	PhysicalTicketStatusVoid       PhysicalTicketStatus = "VOID"
)

type PhysicalTicketQuery struct {
	IDs         []string               `query:"id"`
	EventIDs    []string               `query:"event_id"`
	TicketIDs   []string               `query:"ticket_id"`
	TicketTypes []string               `query:"ticket_type"`
	QRCodes     []string               `query:"qr_code"`
	Statuses    []PhysicalTicketStatus `query:"status"`
}

type PhysicalTicket struct {
	ID      pubEntity.UUID `json:"id"`
	EventID pubEntity.UUID `json:"event_id"`

	TicketType string         `json:"ticket_type"`
	TicketID   pubEntity.UUID `json:"ticket_id"`

	QRCode     string `json:"qr_code"`
	QRCodeHash string `json:"qr_code_hash"`

	Status    PhysicalTicketStatus `json:"status"`
	ScanCount int                  `json:"scan_count"`

	CheckedInAt  *time.Time `json:"checked_in_at"`
	CheckedOutAt *time.Time `json:"checked_out_at"`

	pubEntity.DaoEntity
}

type PhysicalTickets []PhysicalTicket
