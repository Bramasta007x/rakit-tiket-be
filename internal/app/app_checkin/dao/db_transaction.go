package dao

import (
	"context"
	"database/sql"

	"rakit-tiket-be/internal/pkg/dao"
	"rakit-tiket-be/pkg/util"
)

type DBTransaction interface {
	dao.DBTransaction

	GetPhysicalTicketDAO() PhysicalTicketDAO
	GetGateConfigDAO() GateConfigDAO
	GetGateLogDAO() GateLogDAO
}

type dbTransaction struct {
	dao.DBTransaction

	physicalTicketDAO PhysicalTicketDAO
	gateConfigDAO     GateConfigDAO
	gateLogDAO        GateLogDAO
}

func NewTransactionGate(ctx context.Context, log util.LogUtil, sqlDB *sql.DB) DBTransaction {
	dbTrx := &dbTransaction{
		DBTransaction: dao.NewTransaction(ctx, sqlDB),
	}

	dbTrx.physicalTicketDAO = MakePhysicalTicketDAO(log, dbTrx)
	dbTrx.gateConfigDAO = MakeGateConfigDAO(log, dbTrx)
	dbTrx.gateLogDAO = MakeGateLogDAO(log, dbTrx)

	return dbTrx
}

func (dbTrx *dbTransaction) GetPhysicalTicketDAO() PhysicalTicketDAO {
	return dbTrx.physicalTicketDAO
}

func (dbTrx *dbTransaction) GetGateConfigDAO() GateConfigDAO {
	return dbTrx.gateConfigDAO
}

func (dbTrx *dbTransaction) GetGateLogDAO() GateLogDAO {
	return dbTrx.gateLogDAO
}
