package entity

import (
	"time"

	pubEntity "rakit-tiket-be/pkg/entity"
)

type TicketStatus string

const (
	TicketStatusAvailable TicketStatus = "AVAILABLE" // available_qty
	TicketStatusBooked    TicketStatus = "BOOKOUT"   // booked_qty
	TicketStatusSold      TicketStatus = "SOLD"      // sold_qty
)

type (
	TicketQuery struct {
		IDs       []string       `query:"id"`
		EventIDs  []string       `query:"event_id"`
		Types     []string       `query:"type"`
		IsPresale *bool          `query:"is_presale"`
		Statuses  []TicketStatus `query:"status"`
	}

	Ticket struct {
		ID      pubEntity.UUID `json:"id"`
		EventID pubEntity.UUID `json:"event_id"`

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

		// Hype / Urgency (Rakit-Hype)
		SaleStartTime *time.Time `json:"sale_start_time"`
		SaleEndTime   *time.Time `json:"sale_end_time"`

		// Flash Sale
		IsFlashSale    bool       `json:"is_flash_sale"`
		FlashSalePrice *float64   `json:"flash_sale_price"`
		FlashStartTime *time.Time `json:"flash_start_time"`
		FlashEndTime   *time.Time `json:"flash_end_time"`

		// Stock Alert (20% threshold default)
		LowStockThreshold *int `json:"low_stock_threshold"`
		ShowStockAlert    bool `json:"show_stock_alert"`

		// FOMO / Urgency
		ShowCountdown bool       `json:"show_countdown"`
		CountdownEnd  *time.Time `json:"countdown_end"`

		pubEntity.DaoEntity
	}

	// TicketDisplay is computed display info for frontend
	TicketDisplay struct {
		ID                string   `json:"id"`
		EventID           string   `json:"event_id"`
		Type              string   `json:"type"`
		Title             string   `json:"title"`
		Status            string   `json:"status"`
		Total             int      `json:"total"`
		AvailableQty      int      `json:"available_qty"`
		SoldQty           int      `json:"sold_qty"`
		OriginalPrice     float64  `json:"original_price"`
		CurrentPrice      float64  `json:"current_price"`
		IsFlashSale       bool     `json:"is_flash_sale"`
		FlashSalePrice    *float64 `json:"flash_sale_price"`
		FlashEndTime      *string  `json:"flash_end_time"`
		IsOnFlashSale     bool     `json:"is_on_flash_sale"`
		IsLowStock        bool     `json:"is_low_stock"`
		LowStockMessage   *string  `json:"low_stock_message"`
		StockRemaining    int      `json:"stock_remaining"`
		StockPercentage   int      `json:"stock_percentage"`
		IsAvailable       bool     `json:"is_available"`
		UnAvailableReason *string  `json:"unavailable_reason"`
		CountdownSeconds  *int64   `json:"countdown_seconds"`
		CountdownEnd      *string  `json:"countdown_end"`
		ShowCountdown     bool     `json:"show_countdown"`
		SaleStartTime     *string  `json:"sale_start_time"`
		SaleEndTime       *string  `json:"sale_end_time"`
		UrgentMessage     *string  `json:"urgent_message"`
	}

	Tickets []Ticket
)
