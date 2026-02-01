package dao

import (
	"context"
	"database/sql"
)

type DBTransaction interface {
	GetSqlTx() *sql.Tx
	GetSqlDB() *sql.DB
}

type dbTransaction struct {
	sqlDB *sql.DB
	sqlTx *sql.Tx
}

func NewTransaction(ctx context.Context, sqlDB *sql.DB) DBTransaction {
	sqlTx, _ := sqlDB.BeginTx(ctx, nil)
	dbTrx := &dbTransaction{
		sqlTx: sqlTx,
		sqlDB: sqlDB,
	}

	return dbTrx
}

func (dbTrx dbTransaction) GetSqlTx() *sql.Tx {
	return dbTrx.sqlTx
}

func (dbTx dbTransaction) GetSqlDB() *sql.DB {
	return dbTx.sqlDB
}
