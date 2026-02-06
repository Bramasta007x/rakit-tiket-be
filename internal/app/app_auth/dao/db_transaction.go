package dao

import (
	"context"
	"database/sql"
	"rakit-tiket-be/internal/pkg/dao"
)

type DBTransaction interface {
	dao.DBTransaction
	GetUserDAO() UserDAO
}

type dbTransaction struct {
	dao.DBTransaction
	userDAO UserDAO
}

func NewTransaction(ctx context.Context, sqlDB *sql.DB) DBTransaction {
	dbTrx := &dbTransaction{
		DBTransaction: dao.NewTransaction(ctx, sqlDB),
	}
	dbTrx.userDAO = MakeUserDAO(dbTrx)
	return dbTrx
}

func (dbTrx *dbTransaction) GetUserDAO() UserDAO {
	return dbTrx.userDAO
}