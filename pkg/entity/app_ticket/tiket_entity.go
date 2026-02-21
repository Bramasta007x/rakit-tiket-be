package entity

import (
	pubEntity "rakit-tiket-be/pkg/entity"
)

type TicketStatus string

const (
	TicketStatusAvailable TicketStatus = "AVAILABLE"
	TicketStatusBooked    TicketStatus = "BOOKED"
	TicketStatusSold      TicketStatus = "SOLD"
)

type (
	TicketQuery struct {
		IDs       []string       `query:"id"`
		Types     []string       `query:"type"`
		IsPresale *bool          `query:"is_presale"`
		Statuses  []TicketStatus `query:"status"`
	}

	Ticket struct {
		ID pubEntity.UUID `json:"id"`

		// Ticket Info
		Type        string       `json:"type"` // PRESALE_GOLD, GOLD
		Title       string       `json:"title"`
		Status      TicketStatus `json:"status"`
		Description *string      `json:"description"`

		// Pricing & Stock
		Price     float64 `json:"price"`
		Total     int     `json:"total"`     // Total Quantity Tiket
		Remaining int     `json:"remaining"` // Remaining -> Available Qty

		// Flags & Ordering
		IsPresale     bool `json:"is_presale"`
		OrderPriority int  `json:"order_priority"`

		pubEntity.DaoEntity
	}

	Tickets []Ticket
)
