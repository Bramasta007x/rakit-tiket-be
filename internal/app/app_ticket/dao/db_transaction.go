package dao

import (
	"context"
	"database/sql"

	baseDao "rakit-tiket-be/internal/pkg/dao"
	"rakit-tiket-be/pkg/util" // Tambahkan import util
)

type DBTransaction interface {
	baseDao.DBTransaction

	GetTicketDAO() TicketDAO
}

type dbTransaction struct {
	baseDao.DBTransaction

	ticketDAO TicketDAO
}

func NewTransactionTicket(ctx context.Context, log util.LogUtil, sqlDB *sql.DB) DBTransaction {
	dbTrx := &dbTransaction{
		DBTransaction: baseDao.NewTransaction(ctx, sqlDB),
	}

	dbTrx.ticketDAO = MakeTicketDAO(log, dbTrx)
	return dbTrx
}

func (dbTrx *dbTransaction) GetTicketDAO() TicketDAO {
	return dbTrx.ticketDAO
}
