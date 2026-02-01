package service

import (
	"context"
	"database/sql"

	"rakit-tiket-be/internal/app/app_landing_page/dao"
	"rakit-tiket-be/pkg/entity"
)

type LandingPageService interface {
	Search(ctx context.Context, query entity.LandingPageQuery) (entity.LandingPages, error)
	Insert(ctx context.Context, pages entity.LandingPages) error
	Update(ctx context.Context, pages entity.LandingPages) error
	Delete(ctx context.Context, id entity.UUID) error
	SoftDelete(ctx context.Context, id entity.UUID) error
}

type landingPageService struct {
	sqlDB *sql.DB
}

func MakeLandingPageService(sqlDB *sql.DB) LandingPageService {
	return landingPageService{
		sqlDB: sqlDB,
	}
}

func (s landingPageService) Search(
	ctx context.Context,
	query entity.LandingPageQuery,
) (entity.LandingPages, error) {

	dbTrx := dao.NewTransaction(ctx, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	pages, err := dbTrx.GetLandingPageDAO().Search(ctx, query)
	if err != nil {
		return nil, err
	}

	return pages, nil
}

func (s landingPageService) Insert(
	ctx context.Context,
	pages entity.LandingPages,
) error {

	dbTrx := dao.NewTransaction(ctx, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	if err := dbTrx.GetLandingPageDAO().Insert(ctx, pages); err != nil {
		return err
	}

	if err := dbTrx.GetSqlTx().Commit(); err != nil {
		return err
	}

	return nil
}

func (s landingPageService) Update(
	ctx context.Context,
	pages entity.LandingPages,
) error {

	dbTrx := dao.NewTransaction(ctx, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	if err := dbTrx.GetLandingPageDAO().Update(ctx, pages); err != nil {
		return err
	}

	if err := dbTrx.GetSqlTx().Commit(); err != nil {
		return err
	}

	return nil
}

func (s landingPageService) Delete(
	ctx context.Context,
	id entity.UUID,
) error {

	dbTrx := dao.NewTransaction(ctx, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	if err := dbTrx.GetLandingPageDAO().SoftDelete(ctx, id); err != nil {
		return err
	}

	if err := dbTrx.GetSqlTx().Commit(); err != nil {
		return err
	}

	return nil
}

func (s landingPageService) SoftDelete(
	ctx context.Context,
	id entity.UUID,
) error {

	dbTrx := dao.NewTransaction(ctx, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	if err := dbTrx.GetLandingPageDAO().SoftDelete(ctx, id); err != nil {
		return err
	}

	if err := dbTrx.GetSqlTx().Commit(); err != nil {
		return err
	}

	return nil
}
