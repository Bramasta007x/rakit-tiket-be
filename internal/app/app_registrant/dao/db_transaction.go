package dao

import (
	"context"
	"database/sql"

	baseDao "rakit-tiket-be/internal/pkg/dao"
)

type DBTransaction interface {
	baseDao.DBTransaction

	GetRegistrantDAO() RegistrantDAO
	GetAttendeeDAO() AttendeeDAO
}

type dbTransaction struct {
	baseDao.DBTransaction

	registrantDAO RegistrantDAO
	attendeeDAO   AttendeeDAO
}

func NewTransactionRegistrant(ctx context.Context, sqlDB *sql.DB) DBTransaction {
	dbTrx := &dbTransaction{
		DBTransaction: baseDao.NewTransaction(ctx, sqlDB),
	}

	dbTrx.registrantDAO = MakeRegistrantDAO(dbTrx)
	dbTrx.attendeeDAO = MakeAttendeeDAO(dbTrx)
	return dbTrx
}

func (dbTrx *dbTransaction) GetRegistrantDAO() RegistrantDAO {
	return dbTrx.registrantDAO
}

func (dbTrx *dbTransaction) GetAttendeeDAO() AttendeeDAO {
	return dbTrx.attendeeDAO
}
