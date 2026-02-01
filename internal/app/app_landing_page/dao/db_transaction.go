package dao

import (
	"context"
	"database/sql"
	"rakit-tiket-be/internal/pkg/dao"
)

type DBTransaction interface {
	dao.DBTransaction

	GetLandingPageDAO() LandingPageDAO
}

type dbTransaction struct {
	dao.DBTransaction

	organizationDAO LandingPageDAO
}

func NewTransaction(ctx context.Context, sqlDB *sql.DB) DBTransaction {
	dbTrx := &dbTransaction{
		DBTransaction: dao.NewTransaction(ctx, sqlDB),
	}

	dbTrx.organizationDAO = MakeLandingPageDAO(dbTrx)
	return dbTrx
}

func (dbTrx *dbTransaction) GetLandingPageDAO() LandingPageDAO {
	return dbTrx.organizationDAO
}
