package service

import (
	"context"
	"database/sql"

	"rakit-tiket-be/internal/app/app_artist/dao"
	pubEntity "rakit-tiket-be/pkg/entity"
	entity "rakit-tiket-be/pkg/entity/app_artist"
	"rakit-tiket-be/pkg/util"
)

type ArtistService interface {
	Search(ctx context.Context, query entity.ArtistQuery) (entity.Artists, error)
	SearchByID(ctx context.Context, id pubEntity.UUID) (entity.Artist, error)
	Insert(ctx context.Context, artists entity.Artists) error
	Update(ctx context.Context, artists entity.Artists) error
	Delete(ctx context.Context, id pubEntity.UUID) error
	SoftDelete(ctx context.Context, id pubEntity.UUID) error
}

type artistService struct {
	log   util.LogUtil
	sqlDB *sql.DB
}

func MakeArtistService(log util.LogUtil, sqlDB *sql.DB) ArtistService {
	return artistService{
		log:   log,
		sqlDB: sqlDB,
	}
}

func (s artistService) Search(ctx context.Context, query entity.ArtistQuery) (entity.Artists, error) {
	dbTrx := dao.NewTransactionArtist(ctx, s.log, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	artists, err := dbTrx.GetArtistDAO().Search(ctx, query)
	if err != nil {
		return nil, err
	}

	return artists, nil
}

func (s artistService) SearchByID(ctx context.Context, id pubEntity.UUID) (entity.Artist, error) {
	dbTrx := dao.NewTransactionArtist(ctx, s.log, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	artist, err := dbTrx.GetArtistDAO().SearchByID(ctx, id)
	if err != nil {
		return entity.Artist{}, err
	}

	return artist, nil
}

func (s artistService) Insert(ctx context.Context, artists entity.Artists) error {
	dbTrx := dao.NewTransactionArtist(ctx, s.log, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	if err := dbTrx.GetArtistDAO().Insert(ctx, artists); err != nil {
		return err
	}

	if err := dbTrx.GetSqlTx().Commit(); err != nil {
		return err
	}

	return nil
}

func (s artistService) Update(ctx context.Context, artists entity.Artists) error {
	dbTrx := dao.NewTransactionArtist(ctx, s.log, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	if err := dbTrx.GetArtistDAO().Update(ctx, artists); err != nil {
		return err
	}

	if err := dbTrx.GetSqlTx().Commit(); err != nil {
		return err
	}

	return nil
}

func (s artistService) Delete(ctx context.Context, id pubEntity.UUID) error {
	dbTrx := dao.NewTransactionArtist(ctx, s.log, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	if err := dbTrx.GetArtistDAO().Delete(ctx, id); err != nil {
		return err
	}

	if err := dbTrx.GetSqlTx().Commit(); err != nil {
		return err
	}

	return nil
}

func (s artistService) SoftDelete(ctx context.Context, id pubEntity.UUID) error {
	dbTrx := dao.NewTransactionArtist(ctx, s.log, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	if err := dbTrx.GetArtistDAO().SoftDelete(ctx, id); err != nil {
		return err
	}

	if err := dbTrx.GetSqlTx().Commit(); err != nil {
		return err
	}

	return nil
}
