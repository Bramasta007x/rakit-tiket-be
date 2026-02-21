package service

import (
	"context"
	"database/sql"

	"rakit-tiket-be/internal/app/app_ticket/dao"
	pubEntity "rakit-tiket-be/pkg/entity"
	entity "rakit-tiket-be/pkg/entity/app_ticket"
)

type TicketService interface {
	Search(ctx context.Context, query entity.TicketQuery) (entity.Tickets, error)
	Insert(ctx context.Context, tickets entity.Tickets) error
	Update(ctx context.Context, tickets entity.Tickets) error
	Delete(ctx context.Context, id pubEntity.UUID) error
	SoftDelete(ctx context.Context, id pubEntity.UUID) error
}

type ticketService struct {
	sqlDB *sql.DB
}

func MakeTicketService(sqlDB *sql.DB) TicketService {
	return ticketService{
		sqlDB: sqlDB,
	}
}

func (s ticketService) Search(ctx context.Context, query entity.TicketQuery) (entity.Tickets, error) {

	dbTrx := dao.NewTransactionTicket(ctx, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	tickets, err := dbTrx.GetTicketDAO().Search(ctx, query)
	if err != nil {
		return nil, err
	}

	return tickets, nil
}

func (s ticketService) Insert(ctx context.Context, tickets entity.Tickets) error {

	dbTrx := dao.NewTransactionTicket(ctx, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	if err := dbTrx.GetTicketDAO().Insert(ctx, tickets); err != nil {
		return err
	}

	if err := dbTrx.GetSqlTx().Commit(); err != nil {
		return err
	}

	return nil
}

func (s ticketService) Update(ctx context.Context, tickets entity.Tickets) error {

	dbTrx := dao.NewTransactionTicket(ctx, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	if err := dbTrx.GetTicketDAO().Update(ctx, tickets); err != nil {
		return err
	}

	if err := dbTrx.GetSqlTx().Commit(); err != nil {
		return err
	}

	return nil
}

func (s ticketService) Delete(ctx context.Context, id pubEntity.UUID) error {

	dbTrx := dao.NewTransactionTicket(ctx, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	if err := dbTrx.GetTicketDAO().Delete(ctx, id); err != nil {
		return err
	}

	if err := dbTrx.GetSqlTx().Commit(); err != nil {
		return err
	}

	return nil
}

func (s ticketService) SoftDelete(ctx context.Context, id pubEntity.UUID) error {

	dbTrx := dao.NewTransactionTicket(ctx, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	if err := dbTrx.GetTicketDAO().SoftDelete(ctx, id); err != nil {
		return err
	}

	if err := dbTrx.GetSqlTx().Commit(); err != nil {
		return err
	}

	return nil
}
