package dao

import (
	"context"
	"database/sql"

	baseDao "rakit-tiket-be/internal/pkg/dao"
)

type DBTransaction interface {
	baseDao.DBTransaction

	GetTicketDAO() TicketDAO
}

type dbTransaction struct {
	baseDao.DBTransaction

	ticketDAO TicketDAO
}

func NewTransactionTicket(ctx context.Context, sqlDB *sql.DB) DBTransaction {
	dbTrx := &dbTransaction{
		DBTransaction: baseDao.NewTransaction(ctx, sqlDB),
	}

	dbTrx.ticketDAO = MakeTicketDAO(dbTrx)
	return dbTrx
}

func (dbTrx *dbTransaction) GetTicketDAO() TicketDAO {
	return dbTrx.ticketDAO
}
