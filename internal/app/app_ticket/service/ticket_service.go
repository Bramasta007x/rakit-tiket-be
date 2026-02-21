package service

import (
	"context"
	"database/sql"
	"fmt"

	"rakit-tiket-be/internal/app/app_ticket/dao"
	pubEntity "rakit-tiket-be/pkg/entity"
	entity "rakit-tiket-be/pkg/entity/app_ticket"
)

type TicketService interface {
	Search(ctx context.Context, query entity.TicketQuery) (entity.Tickets, error)
	Insert(ctx context.Context, tickets entity.Tickets) error
	Update(ctx context.Context, tickets entity.Tickets) error
	Delete(ctx context.Context, id pubEntity.UUID) error
	SoftDelete(ctx context.Context, id pubEntity.UUID) error
}

type ticketService struct {
	sqlDB *sql.DB
}

func MakeTicketService(sqlDB *sql.DB) TicketService {
	return ticketService{
		sqlDB: sqlDB,
	}
}

func determineTicketStatus(available, booked int) entity.TicketStatus {
	if available > 0 {
		return entity.TicketStatusAvailable
	}

	if available == 0 && booked == 0 {
		return entity.TicketStatusBooked
	}

	return entity.TicketStatusSold
}

func (s ticketService) Search(ctx context.Context, query entity.TicketQuery) (entity.Tickets, error) {

	dbTrx := dao.NewTransactionTicket(ctx, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	tickets, err := dbTrx.GetTicketDAO().Search(ctx, query)
	if err != nil {
		return nil, err
	}

	return tickets, nil
}

func (s ticketService) Insert(ctx context.Context, tickets entity.Tickets) error {

	dbTrx := dao.NewTransactionTicket(ctx, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	// Saat awal dibuat, Available = Total. Sisanya 0.
	for i, t := range tickets {
		t.AvailableQty = t.Total
		t.BookedQty = 0
		t.SoldQty = 0
		t.Status = determineTicketStatus(t.AvailableQty, t.BookedQty)
		tickets[i] = t
	}

	if err := dbTrx.GetTicketDAO().Insert(ctx, tickets); err != nil {
		return err
	}

	if err := dbTrx.GetSqlTx().Commit(); err != nil {
		return err
	}

	return nil
}

func (s ticketService) Update(ctx context.Context, tickets entity.Tickets) error {

	dbTrx := dao.NewTransactionTicket(ctx, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	// 1. Fetch data tiket yang ada (existing) untuk validasi qty
	var ticketIDs []string
	for _, t := range tickets {
		ticketIDs = append(ticketIDs, string(t.ID))
	}

	existingTickets, err := dbTrx.GetTicketDAO().Search(ctx, entity.TicketQuery{IDs: ticketIDs})

	if err != nil {
		return err
	}

	// Mapping existing data agar mudah dicari
	existingMap := make(map[string]entity.Ticket)
	for _, et := range existingTickets {
		existingMap[string(et.ID)] = et
	}

	// 2. Kalkulasi Ulang Stok (Invariant Rule)
	for i, newTicket := range tickets {
		existingData, exists := existingMap[string(newTicket.ID)]
		if !exists {
			return fmt.Errorf("tiket dengan ID %s tidak ditemukan", newTicket.ID)
		}

		// Hitung jumlah tiket yang sudah di luar jangkauan (sudah dibayar / dibooking)
		lockedQty := existingData.BookedQty + existingData.SoldQty

		// Jika admin mengedit total menjadi lebih kecil dari yang sudah laku
		if newTicket.Total < lockedQty {
			return fmt.Errorf("tidak bisa mengurangi total tiket '%s' menjadi %d. Saat ini sudah ada %d tiket yang terjual/dibooking", newTicket.Title, newTicket.Total, lockedQty)
		}

		// Kalkulasi ulang available_qty sesuai rumus: available = total - (booked + sold)
		newTicket.AvailableQty = newTicket.Total - lockedQty

		// Booked dan Sold tidak boleh diedit secara manual melalui form Update Admin
		newTicket.BookedQty = existingData.BookedQty
		newTicket.SoldQty = existingData.SoldQty

		// Update status dinamis
		newTicket.Status = determineTicketStatus(newTicket.AvailableQty, newTicket.BookedQty)

		tickets[i] = newTicket
	}

	if err := dbTrx.GetTicketDAO().Update(ctx, tickets); err != nil {
		return err
	}

	if err := dbTrx.GetSqlTx().Commit(); err != nil {
		return err
	}

	return nil
}

func (s ticketService) Delete(ctx context.Context, id pubEntity.UUID) error {

	dbTrx := dao.NewTransactionTicket(ctx, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	if err := dbTrx.GetTicketDAO().Delete(ctx, id); err != nil {
		return err
	}

	if err := dbTrx.GetSqlTx().Commit(); err != nil {
		return err
	}

	return nil
}

func (s ticketService) SoftDelete(ctx context.Context, id pubEntity.UUID) error {

	dbTrx := dao.NewTransactionTicket(ctx, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	if err := dbTrx.GetTicketDAO().SoftDelete(ctx, id); err != nil {
		return err
	}

	if err := dbTrx.GetSqlTx().Commit(); err != nil {
		return err
	}

	return nil
}
