package dao

import (
	"context"
	"database/sql"

	orderDao "rakit-tiket-be/internal/app/app_order/dao"
	"rakit-tiket-be/internal/pkg/dao"
)

type DBTransaction interface {
	dao.DBTransaction

	GetRegistrantDAO() RegistrantDAO
	GetAttendeeDAO() AttendeeDAO

	GetOrderDAO() orderDao.OrderDAO
}

type dbTransaction struct {
	dao.DBTransaction

	registrantDAO RegistrantDAO
	attendeeDAO   AttendeeDAO
	orderDAO      orderDao.OrderDAO
}

func NewTransactionRegistrant(ctx context.Context, sqlDB *sql.DB) DBTransaction {
	dbTrx := &dbTransaction{
		DBTransaction: dao.NewTransaction(ctx, sqlDB),
	}

	dbTrx.registrantDAO = MakeRegistrantDAO(dbTrx)
	dbTrx.attendeeDAO = MakeAttendeeDAO(dbTrx)
	dbTrx.orderDAO = orderDao.MakeOrderDAO(dbTrx)
	return dbTrx
}

func (dbTrx *dbTransaction) GetRegistrantDAO() RegistrantDAO {
	return dbTrx.registrantDAO
}

func (dbTrx *dbTransaction) GetAttendeeDAO() AttendeeDAO {
	return dbTrx.attendeeDAO
}

func (dbTrx *dbTransaction) GetOrderDAO() orderDao.OrderDAO {
	return dbTrx.orderDAO
}
