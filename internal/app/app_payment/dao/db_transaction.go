package dao

import (
	"context"
	"database/sql"

	eventDao "rakit-tiket-be/internal/app/app_event/dao"
	orderDao "rakit-tiket-be/internal/app/app_order/dao"
	regDao "rakit-tiket-be/internal/app/app_registrant/dao"
	ticketDao "rakit-tiket-be/internal/app/app_ticket/dao"
	"rakit-tiket-be/internal/pkg/dao"
	"rakit-tiket-be/pkg/util"
)

type DBTransaction interface {
	dao.DBTransaction

	GetBankAccountDAO() BankAccountDAO
	GetManualTransferDAO() ManualTransferDAO
	GetGatewayDAO() GatewayDAO
	GetPaymentSettingDAO() PaymentSettingDAO
	GetRegistrantDAO() regDao.RegistrantDAO
	GetAttendeeDAO() regDao.AttendeeDAO
	GetOrderDAO() orderDao.OrderDAO
	GetTicketDAO() ticketDao.TicketDAO
	GetEventDAO() eventDao.EventDAO
}

type dbTransaction struct {
	dao.DBTransaction

	bankAccountDAO    BankAccountDAO
	manualTransferDAO ManualTransferDAO
	gatewayDAO        GatewayDAO
	paymentSettingDAO PaymentSettingDAO
	registrantDAO     regDao.RegistrantDAO
	attendeeDAO       regDao.AttendeeDAO
	orderDAO          orderDao.OrderDAO
	ticketDAO         ticketDao.TicketDAO
	eventDAO          eventDao.EventDAO
}

func NewTransactionPayment(ctx context.Context, log util.LogUtil, sqlDB *sql.DB) DBTransaction {
	dbTrx := &dbTransaction{
		DBTransaction: dao.NewTransaction(ctx, sqlDB),
	}

	dbTrx.bankAccountDAO = MakeBankAccountDAO(log, dbTrx)
	dbTrx.manualTransferDAO = MakeManualTransferDAO(log, dbTrx)
	dbTrx.gatewayDAO = MakeGatewayDAO(log, dbTrx)
	dbTrx.paymentSettingDAO = MakePaymentSettingDAO(log, dbTrx)
	dbTrx.registrantDAO = regDao.MakeRegistrantDAO(log, dbTrx)
	dbTrx.attendeeDAO = regDao.MakeAttendeeDAO(log, dbTrx)
	dbTrx.orderDAO = orderDao.MakeOrderDAO(log, dbTrx)
	dbTrx.ticketDAO = ticketDao.MakeTicketDAO(log, dbTrx)
	dbTrx.eventDAO = eventDao.MakeEventDAO(log, dbTrx)

	return dbTrx
}

func (dbTrx *dbTransaction) GetBankAccountDAO() BankAccountDAO {
	return dbTrx.bankAccountDAO
}

func (dbTrx *dbTransaction) GetManualTransferDAO() ManualTransferDAO {
	return dbTrx.manualTransferDAO
}

func (dbTrx *dbTransaction) GetRegistrantDAO() regDao.RegistrantDAO {
	return dbTrx.registrantDAO
}

func (dbTrx *dbTransaction) GetAttendeeDAO() regDao.AttendeeDAO {
	return dbTrx.attendeeDAO
}

func (dbTrx *dbTransaction) GetOrderDAO() orderDao.OrderDAO {
	return dbTrx.orderDAO
}

func (dbTrx *dbTransaction) GetTicketDAO() ticketDao.TicketDAO {
	return dbTrx.ticketDAO
}

func (dbTrx *dbTransaction) GetEventDAO() eventDao.EventDAO {
	return dbTrx.eventDAO
}

func (dbTrx *dbTransaction) GetGatewayDAO() GatewayDAO {
	return dbTrx.gatewayDAO
}

func (dbTrx *dbTransaction) GetPaymentSettingDAO() PaymentSettingDAO {
	return dbTrx.paymentSettingDAO
}
