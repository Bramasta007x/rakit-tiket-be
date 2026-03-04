package service

import (
	"context"
	"database/sql"

	"rakit-tiket-be/internal/app/app_event/dao"
	pubEntity "rakit-tiket-be/pkg/entity"
	entity "rakit-tiket-be/pkg/entity/app_event"
	"rakit-tiket-be/pkg/util"
)

type EventService interface {
	Search(ctx context.Context, query entity.EventQuery) (entity.Events, error)
	Insert(ctx context.Context, events entity.Events) error
	Update(ctx context.Context, events entity.Events) error
	Delete(ctx context.Context, id pubEntity.UUID) error
	SoftDelete(ctx context.Context, id pubEntity.UUID) error
}

type eventService struct {
	log   util.LogUtil
	sqlDB *sql.DB
}

func MakeEventService(log util.LogUtil, sqlDB *sql.DB) EventService {
	return eventService{
		log:   log,
		sqlDB: sqlDB,
	}
}

func (s eventService) Search(ctx context.Context, query entity.EventQuery) (entity.Events, error) {
	// Menggunakan NewTransactionEvent yang kita buat sebelumnya
	dbTrx := dao.NewTransactionEvent(ctx, s.log, s.sqlDB)
	// Kita tidak Commit di Search karena hanya operasi Read
	defer dbTrx.GetSqlTx().Rollback()

	events, err := dbTrx.GetEventDAO().Search(ctx, query)
	if err != nil {
		return nil, err
	}

	return events, nil
}

func (s eventService) Insert(ctx context.Context, events entity.Events) error {
	dbTrx := dao.NewTransactionEvent(ctx, s.log, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	if err := dbTrx.GetEventDAO().Insert(ctx, events); err != nil {
		return err
	}

	if err := dbTrx.GetSqlTx().Commit(); err != nil {
		return err
	}

	return nil
}

func (s eventService) Update(ctx context.Context, events entity.Events) error {
	dbTrx := dao.NewTransactionEvent(ctx, s.log, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	if err := dbTrx.GetEventDAO().Update(ctx, events); err != nil {
		return err
	}

	if err := dbTrx.GetSqlTx().Commit(); err != nil {
		return err
	}

	return nil
}

func (s eventService) Delete(ctx context.Context, id pubEntity.UUID) error {
	dbTrx := dao.NewTransactionEvent(ctx, s.log, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	// Secara permanen menghapus data dari database
	if err := dbTrx.GetEventDAO().Delete(ctx, id); err != nil {
		return err
	}

	if err := dbTrx.GetSqlTx().Commit(); err != nil {
		return err
	}

	return nil
}

func (s eventService) SoftDelete(ctx context.Context, id pubEntity.UUID) error {
	dbTrx := dao.NewTransactionEvent(ctx, s.log, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	// Hanya mengubah flag 'deleted' menjadi true
	if err := dbTrx.GetEventDAO().SoftDelete(ctx, id); err != nil {
		return err
	}

	if err := dbTrx.GetSqlTx().Commit(); err != nil {
		return err
	}

	return nil
}
