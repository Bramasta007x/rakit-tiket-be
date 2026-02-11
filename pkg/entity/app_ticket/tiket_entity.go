package entity

import (
	pubEntity "rakit-tiket-be/pkg/entity"
)

type (
	TicketQuery struct {
		IDs       []string `query:"id"`
		Types     []string `query:"type"`
		IsPresale *bool    `query:"is_presale"`
	}

	Ticket struct {
		ID pubEntity.UUID `json:"id"`

		// Ticket Info
		Type        string  `json:"type"` // PRESALE_GOLD, GOLD
		Title       string  `json:"title"`
		Description *string `json:"description"`

		// Pricing & Stock
		Price     float64 `json:"price"`
		Total     int     `json:"total"`
		Remaining int     `json:"remaining"`

		// Flags & Ordering
		IsPresale     bool `json:"is_presale"`
		OrderPriority int  `json:"order_priority"`

		pubEntity.DaoEntity
	}

	Tickets []Ticket
)
