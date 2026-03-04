package dao

import (
	"context"
	"database/sql"

	baseDao "rakit-tiket-be/internal/pkg/dao"
	"rakit-tiket-be/pkg/util"
)

type DBTransaction interface {
	baseDao.DBTransaction

	GetEventDAO() EventDAO
}

type dbTransaction struct {
	baseDao.DBTransaction

	eventDAO EventDAO
}

// NewTransactionEvent menginisialisasi transaksi database khusus untuk domain Event
func NewTransactionEvent(ctx context.Context, log util.LogUtil, sqlDB *sql.DB) DBTransaction {
	dbTrx := &dbTransaction{
		DBTransaction: baseDao.NewTransaction(ctx, sqlDB),
	}

	// Inisialisasi EventDAO dengan menyuntikkan dbTrx itu sendiri
	dbTrx.eventDAO = MakeEventDAO(log, dbTrx)
	return dbTrx
}

func (dbTrx *dbTransaction) GetEventDAO() EventDAO {
	return dbTrx.eventDAO
}
