package entity

import (
	"time"

	pubEntity "rakit-tiket-be/pkg/entity"
)

type (
	HypeQuery struct {
		TicketIDs []string `query:"ticket_id"`
		EventIDs  []string `query:"event_id"`
	}

	TicketsHype []TicketHype

	TicketHype struct {
		ID                pubEntity.UUID `json:"id"`
		TicketID          pubEntity.UUID `json:"ticket_id"`
		EventID           pubEntity.UUID `json:"event_id"`
		FlashSalePrice    *float64       `json:"flash_sale_price"`
		FlashStartTime    *time.Time     `json:"flash_start_time"`
		FlashEndTime      *time.Time     `json:"flash_end_time"`
		LowStockThreshold *int           `json:"low_stock_threshold"`
		ShowStockAlert    bool           `json:"show_stock_alert"`
		CountdownEnd      *time.Time     `json:"countdown_end"`
		ShowCountdown     bool           `json:"show_countdown"`

		pubEntity.DaoEntity
	}
)
