package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	ticketDao "rakit-tiket-be/internal/app/app_ticket/dao"
	ticketEntity "rakit-tiket-be/pkg/entity/app_ticket"
	"rakit-tiket-be/pkg/util"
)

type HypeService interface {
	GetActiveTickets(ctx context.Context, eventID string) ([]ticketEntity.TicketDisplay, error)
	GetTicketDisplay(ctx context.Context, ticketID string) (*ticketEntity.TicketDisplay, error)
	CheckAvailability(ctx context.Context, ticketID string) (*AvailabilityCheck, error)

	SetFlashSale(ctx context.Context, req SetFlashSaleRequest) error
	DisableFlashSale(ctx context.Context, ticketID string) error
	SetCountdown(ctx context.Context, req SetCountdownRequest) error
	SetStockAlert(ctx context.Context, req SetStockAlertRequest) error
}

type AvailabilityCheck struct {
	Available      bool     `json:"available"`
	Reason         string   `json:"reason"`
	CurrentPrice   float64  `json:"current_price"`
	IsFlashSale    bool     `json:"is_flash_sale"`
	FlashSalePrice *float64 `json:"flash_sale_price"`
}

type SetFlashSaleRequest struct {
	TicketID   string     `json:"ticket_id"`
	FlashPrice float64    `json:"flash_price"`
	StartTime  *time.Time `json:"start_time"`
	EndTime    *time.Time `json:"end_time"`
}

type SetCountdownRequest struct {
	TicketID string     `json:"ticket_id"`
	EndTime  *time.Time `json:"end_time"`
}

type SetStockAlertRequest struct {
	TicketID          string `json:"ticket_id"`
	LowStockThreshold *int   `json:"low_stock_threshold"`
	ShowStockAlert    bool   `json:"show_stock_alert"`
}

type hypeService struct {
	log   util.LogUtil
	sqlDB *sql.DB
}

func MakeHypeService(log util.LogUtil, sqlDB *sql.DB) HypeService {
	return &hypeService{
		log:   log,
		sqlDB: sqlDB,
	}
}

func (s *hypeService) GetActiveTickets(ctx context.Context, eventID string) ([]ticketEntity.TicketDisplay, error) {
	dbTrx := ticketDao.NewTransactionTicket(ctx, s.log, s.sqlDB)

	tickets, err := dbTrx.GetTicketDAO().Search(ctx, ticketEntity.TicketQuery{
		EventIDs: []string{eventID},
	})
	if err != nil {
		return nil, err
	}

	var result []ticketEntity.TicketDisplay
	for _, t := range tickets {
		display := s.computeDisplay(t)
		result = append(result, display)
	}

	return result, nil
}

func (s *hypeService) GetTicketDisplay(ctx context.Context, ticketID string) (*ticketEntity.TicketDisplay, error) {
	dbTrx := ticketDao.NewTransactionTicket(ctx, s.log, s.sqlDB)

	tickets, err := dbTrx.GetTicketDAO().Search(ctx, ticketEntity.TicketQuery{
		IDs: []string{ticketID},
	})
	if err != nil {
		return nil, err
	}

	if len(tickets) == 0 {
		return nil, fmt.Errorf("ticket tidak ditemukan")
	}

	display := s.computeDisplay(tickets[0])
	return &display, nil
}

func (s *hypeService) CheckAvailability(ctx context.Context, ticketID string) (*AvailabilityCheck, error) {
	dbTrx := ticketDao.NewTransactionTicket(ctx, s.log, s.sqlDB)

	tickets, err := dbTrx.GetTicketDAO().Search(ctx, ticketEntity.TicketQuery{
		IDs: []string{ticketID},
	})
	if err != nil {
		return nil, err
	}

	if len(tickets) == 0 {
		return nil, fmt.Errorf("ticket tidak ditemukan")
	}

	ticket := tickets[0]
	result := &AvailabilityCheck{
		Available:    true,
		CurrentPrice: ticket.Price,
		IsFlashSale:  ticket.IsFlashSale,
	}

	if ticket.FlashSalePrice != nil {
		result.FlashSalePrice = ticket.FlashSalePrice
	}

	now := time.Now()

	if ticket.AvailableQty <= 0 {
		result.Available = false
		result.Reason = "SOLD_OUT"
		return result, nil
	}

	if ticket.SaleStartTime != nil && now.Before(*ticket.SaleStartTime) {
		result.Available = false
		result.Reason = "SALE_NOT_STARTED"
		return result, nil
	}

	if ticket.SaleEndTime != nil && now.After(*ticket.SaleEndTime) {
		result.Available = false
		result.Reason = "SALE_ENDED"
		return result, nil
	}

	if ticket.IsFlashSale && ticket.FlashStartTime != nil && ticket.FlashEndTime != nil {
		if now.Before(*ticket.FlashStartTime) || now.After(*ticket.FlashEndTime) {
			result.CurrentPrice = ticket.Price
			result.IsFlashSale = false
		} else if ticket.FlashSalePrice != nil {
			result.CurrentPrice = *ticket.FlashSalePrice
		}
	}

	return result, nil
}

func (s *hypeService) computeDisplay(t ticketEntity.Ticket) ticketEntity.TicketDisplay {
	now := time.Now()

	display := ticketEntity.TicketDisplay{
		ID:             string(t.ID),
		EventID:        string(t.EventID),
		Type:           t.Type,
		Title:          t.Title,
		Status:         string(t.Status),
		Total:          t.Total,
		AvailableQty:   t.AvailableQty,
		SoldQty:        t.SoldQty,
		OriginalPrice:  t.Price,
		CurrentPrice:   t.Price,
		IsFlashSale:    t.IsFlashSale,
		StockRemaining: t.AvailableQty,
		ShowCountdown:  t.ShowCountdown,
	}

	if t.FlashSalePrice != nil {
		display.FlashSalePrice = t.FlashSalePrice
	}
	if t.FlashEndTime != nil {
		formatted := t.FlashEndTime.Format("2006-01-02T15:04:05Z07:00")
		display.FlashEndTime = &formatted
	}

	isOnFlashSale := t.IsFlashSale && t.FlashStartTime != nil && t.FlashEndTime != nil &&
		now.After(*t.FlashStartTime) && now.Before(*t.FlashEndTime)
	display.IsOnFlashSale = isOnFlashSale

	if isOnFlashSale && t.FlashSalePrice != nil {
		display.CurrentPrice = *t.FlashSalePrice
	}

	threshold := 20
	if t.LowStockThreshold != nil {
		threshold = *t.LowStockThreshold
	}

	if t.ShowStockAlert && t.AvailableQty > 0 {
		percentage := (t.AvailableQty * 100) / t.Total
		if percentage <= threshold {
			display.IsLowStock = true
			msg := fmt.Sprintf("Sisa %d tiket!", t.AvailableQty)
			display.LowStockMessage = &msg
			display.StockPercentage = percentage
		} else {
			display.StockPercentage = percentage
		}
	} else {
		display.StockPercentage = (t.AvailableQty * 100) / t.Total
	}

	if t.SaleStartTime != nil {
		formatted := t.SaleStartTime.Format("2006-01-02T15:04:05Z07:00")
		display.SaleStartTime = &formatted
	}
	if t.SaleEndTime != nil {
		formatted := t.SaleEndTime.Format("2006-01-02T15:04:05Z07:00")
		display.SaleEndTime = &formatted
	}

	display.IsAvailable = true
	if t.AvailableQty <= 0 {
		display.IsAvailable = false
		display.UnAvailableReason = ptrString("SOLD_OUT")
	} else if t.SaleStartTime != nil && now.Before(*t.SaleStartTime) {
		display.IsAvailable = false
		display.UnAvailableReason = ptrString("SALE_NOT_STARTED")
	} else if t.SaleEndTime != nil && now.After(*t.SaleEndTime) {
		display.IsAvailable = false
		display.UnAvailableReason = ptrString("SALE_ENDED")
	}

	if display.IsAvailable && t.ShowCountdown {
		endTime := t.SaleEndTime
		if t.CountdownEnd != nil {
			endTime = t.CountdownEnd
		}

		if endTime != nil {
			countdown := int64(endTime.Unix() - now.Unix())
			if countdown > 0 {
				display.CountdownSeconds = &countdown
				formatted := endTime.Format("2006-01-02T15:04:05Z07:00")
				display.CountdownEnd = &formatted
			}
		}
	}

	if display.IsLowStock {
		display.UrgentMessage = display.LowStockMessage
	} else if isOnFlashSale {
		msg := "Flash sale! Harga spesial!"
		display.UrgentMessage = &msg
	} else if display.CountdownSeconds != nil && *display.CountdownSeconds < 3600 {
		msg := "Segera berakhir!"
		display.UrgentMessage = &msg
	}

	return display
}

func ptrString(s string) *string {
	return &s
}

func (s *hypeService) SetFlashSale(ctx context.Context, req SetFlashSaleRequest) error {
	dbTrx := ticketDao.NewTransactionTicket(ctx, s.log, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	tickets, err := dbTrx.GetTicketDAO().SearchForUpdate(ctx, ticketEntity.TicketQuery{
		IDs: []string{req.TicketID},
	})
	if err != nil {
		return err
	}

	if len(tickets) == 0 {
		return fmt.Errorf("ticket tidak ditemukan")
	}

	ticket := tickets[0]
	ticket.IsFlashSale = true
	ticket.FlashSalePrice = &req.FlashPrice
	ticket.FlashStartTime = req.StartTime
	ticket.FlashEndTime = req.EndTime

	if err := dbTrx.GetTicketDAO().Update(ctx, ticketEntity.Tickets{ticket}); err != nil {
		return err
	}

	return dbTrx.GetSqlTx().Commit()
}

func (s *hypeService) DisableFlashSale(ctx context.Context, ticketID string) error {
	dbTrx := ticketDao.NewTransactionTicket(ctx, s.log, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	tickets, err := dbTrx.GetTicketDAO().SearchForUpdate(ctx, ticketEntity.TicketQuery{
		IDs: []string{ticketID},
	})
	if err != nil {
		return err
	}

	if len(tickets) == 0 {
		return fmt.Errorf("ticket tidak ditemukan")
	}

	ticket := tickets[0]
	ticket.IsFlashSale = false
	ticket.FlashSalePrice = nil
	ticket.FlashStartTime = nil
	ticket.FlashEndTime = nil

	if err := dbTrx.GetTicketDAO().Update(ctx, ticketEntity.Tickets{ticket}); err != nil {
		return err
	}

	return dbTrx.GetSqlTx().Commit()
}

func (s *hypeService) SetCountdown(ctx context.Context, req SetCountdownRequest) error {
	dbTrx := ticketDao.NewTransactionTicket(ctx, s.log, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	tickets, err := dbTrx.GetTicketDAO().SearchForUpdate(ctx, ticketEntity.TicketQuery{
		IDs: []string{req.TicketID},
	})
	if err != nil {
		return err
	}

	if len(tickets) == 0 {
		return fmt.Errorf("ticket tidak ditemukan")
	}

	ticket := tickets[0]
	ticket.CountdownEnd = req.EndTime
	ticket.ShowCountdown = true

	if err := dbTrx.GetTicketDAO().Update(ctx, ticketEntity.Tickets{ticket}); err != nil {
		return err
	}

	return dbTrx.GetSqlTx().Commit()
}

func (s *hypeService) SetStockAlert(ctx context.Context, req SetStockAlertRequest) error {
	dbTrx := ticketDao.NewTransactionTicket(ctx, s.log, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	tickets, err := dbTrx.GetTicketDAO().SearchForUpdate(ctx, ticketEntity.TicketQuery{
		IDs: []string{req.TicketID},
	})
	if err != nil {
		return err
	}

	if len(tickets) == 0 {
		return fmt.Errorf("ticket tidak ditemukan")
	}

	ticket := tickets[0]
	ticket.LowStockThreshold = req.LowStockThreshold
	ticket.ShowStockAlert = req.ShowStockAlert

	if err := dbTrx.GetTicketDAO().Update(ctx, ticketEntity.Tickets{ticket}); err != nil {
		return err
	}

	return dbTrx.GetSqlTx().Commit()
}
