package dao

import (
	"context"
	"fmt"
	"time"

	pubEntity "rakit-tiket-be/pkg/entity"
	entity "rakit-tiket-be/pkg/entity/app_ticket"

	"gitlab.com/threetopia/sqlgo/v2"
)

type TicketDAO interface {
	Search(ctx context.Context, query entity.TicketQuery) (entity.Tickets, error)
	Insert(ctx context.Context, tickets entity.Tickets) error
	Update(ctx context.Context, tickets entity.Tickets) error
	Delete(ctx context.Context, id pubEntity.UUID) error
	SoftDelete(ctx context.Context, id pubEntity.UUID) error
}

type ticketDAO struct {
	dbTrx DBTransaction
}

func MakeTicketDAO(dbTrx DBTransaction) TicketDAO {
	return ticketDAO{
		dbTrx: dbTrx,
	}
}

func (d ticketDAO) Search(ctx context.Context, query entity.TicketQuery) (entity.Tickets, error) {

	sqlSelect := sqlgo.NewSQLGoSelect().
		SetSQLSelect("t.id", "id").
		SetSQLSelect("t.type", "type").
		SetSQLSelect("t.title", "title").
		SetSQLSelect("t.description", "description").
		SetSQLSelect("t.price", "price").
		SetSQLSelect("t.total", "total").
		SetSQLSelect("t.remaining", "remaining").
		SetSQLSelect("t.is_presale", "is_presale").
		SetSQLSelect("t.order_priority", "order_priority").
		SetSQLSelect("t.created_at", "created_at").
		SetSQLSelect("t.updated_at", "updated_at")

	sqlFrom := sqlgo.NewSQLGoFrom().
		SetSQLFrom("tickets", "t")

	sqlWhere := sqlgo.NewSQLGoWhere()

	if len(query.IDs) > 0 {
		sqlWhere.SetSQLWhere("AND", "t.id", "IN", query.IDs)
	}

	if len(query.Types) > 0 {
		sqlWhere.SetSQLWhere("AND", "t.type", "IN", query.Types)
	}

	if query.IsPresale != nil {
		sqlWhere.SetSQLWhere("AND", "t.is_presale", "=", *query.IsPresale)
	}

	sql := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLGoSelect(sqlSelect).
		SetSQLGoFrom(sqlFrom).
		SetSQLGoWhere(sqlWhere)

	rows, err := d.dbTrx.GetSqlDB().QueryContext(
		ctx,
		sql.BuildSQL(),
		sql.GetSQLGoParameter().GetSQLParameter()...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets entity.Tickets
	for rows.Next() {
		var ticket entity.Ticket

		if err := rows.Scan(
			&ticket.ID,
			&ticket.Type,
			&ticket.Title,
			&ticket.Description,
			&ticket.Price,
			&ticket.Total,
			&ticket.Remaining,
			&ticket.IsPresale,
			&ticket.OrderPriority,
			&ticket.CreatedAt,
			&ticket.UpdatedAt,
		); err != nil {
			return nil, err
		}

		tickets = append(tickets, ticket)
	}

	return tickets, nil
}

func (d ticketDAO) Insert(ctx context.Context, tickets entity.Tickets) error {

	if len(tickets) < 1 {
		return fmt.Errorf("empty ticket data")
	}

	sqlInsert := sqlgo.NewSQLGoInsert().
		SetSQLInsert("tickets").
		SetSQLInsertColumn(
			"id",
			"type",
			"title",
			"description",
			"price",
			"total",
			"remaining",
			"is_presale",
			"order_priority",
			"created_at",
		)

	for i, ticket := range tickets {
		ticket.CreatedAt = time.Now()

		if ticket.ID == "" {
			ticket.ID = pubEntity.MakeUUID(
				ticket.Type,
				ticket.CreatedAt.String(),
			)
		}

		sqlInsert.SetSQLInsertValue(
			ticket.ID,
			ticket.Type,
			ticket.Title,
			ticket.Description,
			ticket.Price,
			ticket.Total,
			ticket.Remaining,
			ticket.IsPresale,
			ticket.OrderPriority,
			ticket.CreatedAt,
		)

		tickets[i] = ticket
	}

	sql := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLGoInsert(sqlInsert)

	_, err := d.dbTrx.GetSqlTx().ExecContext(
		ctx,
		sql.BuildSQL(),
		sql.GetSQLGoParameter().GetSQLParameter()...,
	)

	return err
}

func (d ticketDAO) Update(ctx context.Context, tickets entity.Tickets) error {

	if len(tickets) < 1 {
		return fmt.Errorf("empty ticket data")
	}

	for i, ticket := range tickets {
		now := time.Now()
		ticket.UpdatedAt = &now

		sql := sqlgo.NewSQLGo().
			SetSQLSchema("public").
			SetSQLUpdate("tickets").
			SetSQLUpdateValue("type", ticket.Type).
			SetSQLUpdateValue("title", ticket.Title).
			SetSQLUpdateValue("description", ticket.Description).
			SetSQLUpdateValue("price", ticket.Price).
			SetSQLUpdateValue("total", ticket.Total).
			SetSQLUpdateValue("remaining", ticket.Remaining).
			SetSQLUpdateValue("is_presale", ticket.IsPresale).
			SetSQLUpdateValue("order_priority", ticket.OrderPriority).
			SetSQLUpdateValue("updated_at", ticket.UpdatedAt).
			SetSQLWhere("AND", "id", "=", ticket.ID)

		_, err := d.dbTrx.GetSqlTx().ExecContext(
			ctx,
			sql.BuildSQL(),
			sql.GetSQLGoParameter().GetSQLParameter()...,
		)
		if err != nil {
			return err
		}

		tickets[i] = ticket
	}

	return nil
}

func (d ticketDAO) Delete(ctx context.Context, id pubEntity.UUID) error {

	sql := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLDelete("tickets").
		SetSQLWhere("AND", "id", "=", id)

	_, err := d.dbTrx.GetSqlTx().ExecContext(
		ctx,
		sql.BuildSQL(),
		sql.GetSQLGoParameter().GetSQLParameter()...,
	)

	return err
}

func (d ticketDAO) SoftDelete(ctx context.Context, id pubEntity.UUID) error {
	sql := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLUpdate("ticket").
		SetSQLUpdateValue("deleted", true).
		SetSQLWhere("AND", "id", "=", id)

	_, err := d.dbTrx.GetSqlTx().ExecContext(
		ctx,
		sql.BuildSQL(),
		sql.GetSQLGoParameter().GetSQLParameter()...,
	)

	return err
}
