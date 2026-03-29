package dao

import (
	"context"
	"database/sql"

	baseDao "rakit-tiket-be/internal/pkg/dao"
	"rakit-tiket-be/pkg/util"
)

type DBTransaction interface {
	baseDao.DBTransaction

	GetArtistDAO() ArtistDAO
}

type dbTransaction struct {
	baseDao.DBTransaction

	artistDAO ArtistDAO
}

func NewTransactionArtist(ctx context.Context, log util.LogUtil, sqlDB *sql.DB) DBTransaction {
	dbTrx := &dbTransaction{
		DBTransaction: baseDao.NewTransaction(ctx, sqlDB),
	}

	dbTrx.artistDAO = MakeArtistDAO(log, dbTrx)
	return dbTrx
}

func (dbTrx *dbTransaction) GetArtistDAO() ArtistDAO {
	return dbTrx.artistDAO
}
