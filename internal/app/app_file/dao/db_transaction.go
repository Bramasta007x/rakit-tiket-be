package dao

import (
	"context"
	"database/sql"
	"rakit-tiket-be/internal/pkg/dao"
)

type DBTransaction interface {
	dao.DBTransaction

	GetFileDao() FileDAO
}

type dbTransaction struct {
	dao.DBTransaction

	fileDao FileDAO
}

func NewTransaction(ctx context.Context, sqlDB *sql.DB) DBTransaction {
	dbTrx := &dbTransaction{
		DBTransaction: dao.NewTransaction(ctx, sqlDB),
	}

	dbTrx.fileDao = MakeFileDAO(dbTrx)
	return dbTrx
}

func (dbTrx *dbTransaction) GetFileDao() FileDAO {
	return dbTrx.fileDao
}
