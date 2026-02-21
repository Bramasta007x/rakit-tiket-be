package entity

import (
	pubEntity "rakit-tiket-be/pkg/entity"
)

type TicketStatus string

const (
	TicketStatusAvailable TicketStatus = "AVAILABLE" // available_qty
	TicketStatusBooked    TicketStatus = "BOOKED"    // booked_qty
	TicketStatusSold      TicketStatus = "SOLD"      // sold_qty
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
		Price        float64 `json:"price"`
		Total        int     `json:"total"`
		AvailableQty int     `json:"available_qty"`
		BookedQty    int     `json:"booked_qty"`
		SoldQty      int     `json:"sold_qty"`

		/*
			Stock Distribution:
			Available Qty -> available_qty
			Booked Qty    -> booked_qty
			Sold Qty      -> sold_qty

			Rule:
			Available + Booked + Sold = Total
		*/

		// Flags & Ordering
		IsPresale     bool `json:"is_presale"`
		OrderPriority int  `json:"order_priority"`

		pubEntity.DaoEntity
	}

	Tickets []Ticket
)
