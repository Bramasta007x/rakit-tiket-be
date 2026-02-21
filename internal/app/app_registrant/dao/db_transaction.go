package dao

import (
	"context"
	"database/sql"

	orderDao "rakit-tiket-be/internal/app/app_order/dao"
	ticketDao "rakit-tiket-be/internal/app/app_ticket/dao"
	"rakit-tiket-be/internal/pkg/dao"
)

type DBTransaction interface {
	dao.DBTransaction

	GetRegistrantDAO() RegistrantDAO
	GetAttendeeDAO() AttendeeDAO

	GetOrderDAO() orderDao.OrderDAO
	GetTicketDAO() ticketDao.TicketDAO
}

type dbTransaction struct {
	dao.DBTransaction

	registrantDAO RegistrantDAO
	attendeeDAO   AttendeeDAO
	orderDAO      orderDao.OrderDAO
	ticketDAO     ticketDao.TicketDAO
}

func NewTransactionRegistrant(ctx context.Context, sqlDB *sql.DB) DBTransaction {
	dbTrx := &dbTransaction{
		DBTransaction: dao.NewTransaction(ctx, sqlDB),
	}

	dbTrx.registrantDAO = MakeRegistrantDAO(dbTrx)
	dbTrx.attendeeDAO = MakeAttendeeDAO(dbTrx)
	dbTrx.orderDAO = orderDao.MakeOrderDAO(dbTrx)
	dbTrx.ticketDAO = ticketDao.MakeTicketDAO(dbTrx)
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

func (dbTrx *dbTransaction) GetTicketDAO() ticketDao.TicketDAO {
	return dbTrx.ticketDAO
}
