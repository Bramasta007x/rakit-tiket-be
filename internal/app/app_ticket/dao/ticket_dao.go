package dao

import (
	"context"
	"fmt"
	"time"

	pubEntity "rakit-tiket-be/pkg/entity"
	entity "rakit-tiket-be/pkg/entity/app_ticket"
	"rakit-tiket-be/pkg/util" // Import util

	"gitlab.com/threetopia/sqlgo/v2"
	"go.uber.org/zap" // Import zap
)

type TicketDAO interface {
	Search(ctx context.Context, query entity.TicketQuery) (entity.Tickets, error)
	Insert(ctx context.Context, tickets entity.Tickets) error
	Update(ctx context.Context, tickets entity.Tickets) error
	Delete(ctx context.Context, id pubEntity.UUID) error
	SoftDelete(ctx context.Context, id pubEntity.UUID) error
	BookStock(ctx context.Context, id pubEntity.UUID, qty int) error
	ConfirmSold(ctx context.Context, id pubEntity.UUID, qty int) error
	ReleaseBooked(ctx context.Context, id pubEntity.UUID, qty int) error
}

type ticketDAO struct {
	log   util.LogUtil // Tambahkan log util
	dbTrx DBTransaction
}

func MakeTicketDAO(log util.LogUtil, dbTrx DBTransaction) TicketDAO {
	return ticketDAO{
		log:   log,
		dbTrx: dbTrx,
	}
}

func (d ticketDAO) Search(ctx context.Context, query entity.TicketQuery) (entity.Tickets, error) {

	sqlSelect := sqlgo.NewSQLGoSelect().
		SetSQLSelect("t.id", "id").
		SetSQLSelect("t.type", "type").
		SetSQLSelect("t.title", "title").
		SetSQLSelect("t.status", "status").
		SetSQLSelect("t.description", "description").
		SetSQLSelect("t.price", "price").
		SetSQLSelect("t.total", "total").
		SetSQLSelect("t.available_qty", "available_qty").
		SetSQLSelect("t.booked_qty", "booked_qty").
		SetSQLSelect("t.sold_qty", "sold_qty").
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
	if query.Statuses != nil {
		sqlWhere.SetSQLWhere("AND", "t.status", "=", query.Statuses)
	}

	sql := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLGoSelect(sqlSelect).
		SetSQLGoFrom(sqlFrom).
		SetSQLGoWhere(sqlWhere)

	sqlStr := sql.BuildSQL()
	sqlParams := sql.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "ticketDAO.Search",
		zap.String("SQL", sqlStr),
		zap.Any("Params", sqlParams),
	)

	rows, err := d.dbTrx.GetSqlDB().QueryContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		d.log.Error(ctx, "ticketDAO.Search", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var tickets entity.Tickets
	for rows.Next() {
		var ticket entity.Ticket
		if err := rows.Scan(
			&ticket.ID, &ticket.Type, &ticket.Title, &ticket.Status,
			&ticket.Description, &ticket.Price, &ticket.Total,
			&ticket.AvailableQty, &ticket.BookedQty, &ticket.SoldQty,
			&ticket.IsPresale, &ticket.OrderPriority,
			&ticket.CreatedAt, &ticket.UpdatedAt,
		); err != nil {
			d.log.Error(ctx, "ticketDAO.Search.Scan", zap.Error(err))
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
			"id", "type", "title", "status", "description", "price", "total",
			"available_qty", "booked_qty", "sold_qty", "is_presale",
			"order_priority", "created_at",
		)

	for i, ticket := range tickets {
		ticket.CreatedAt = time.Now()
		if ticket.ID == "" {
			ticket.ID = pubEntity.MakeUUID(ticket.Type, ticket.CreatedAt.String())
		}
		sqlInsert.SetSQLInsertValue(
			ticket.ID, ticket.Type, ticket.Title, ticket.Status, ticket.Description,
			ticket.Price, ticket.Total, ticket.AvailableQty, ticket.BookedQty,
			ticket.SoldQty, ticket.IsPresale, ticket.OrderPriority, ticket.CreatedAt,
		)
		tickets[i] = ticket
	}

	sql := sqlgo.NewSQLGo().SetSQLSchema("public").SetSQLGoInsert(sqlInsert)
	sqlStr := sql.BuildSQL()
	sqlParams := sql.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "ticketDAO.Insert", zap.String("SQL", sqlStr), zap.Int("Count", len(tickets)))

	_, err := d.dbTrx.GetSqlTx().ExecContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		d.log.Error(ctx, "ticketDAO.Insert", zap.Error(err))
		return err
	}
	return nil
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
			SetSQLUpdateValue("status", ticket.Status).
			SetSQLUpdateValue("description", ticket.Description).
			SetSQLUpdateValue("price", ticket.Price).
			SetSQLUpdateValue("total", ticket.Total).
			SetSQLUpdateValue("available_qty", ticket.AvailableQty).
			SetSQLUpdateValue("booked_qty", ticket.BookedQty).
			SetSQLUpdateValue("sold_qty", ticket.SoldQty).
			SetSQLUpdateValue("is_presale", ticket.IsPresale).
			SetSQLUpdateValue("order_priority", ticket.OrderPriority).
			SetSQLUpdateValue("updated_at", ticket.UpdatedAt).
			SetSQLWhere("AND", "id", "=", ticket.ID)

		sqlStr := sql.BuildSQL()
		sqlParams := sql.GetSQLGoParameter().GetSQLParameter()

		d.log.Debug(ctx, "ticketDAO.Update", zap.String("ID", string(ticket.ID)))

		_, err := d.dbTrx.GetSqlTx().ExecContext(ctx, sqlStr, sqlParams...)
		if err != nil {
			d.log.Error(ctx, "ticketDAO.Update", zap.Error(err))
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

	sqlStr := sql.BuildSQL()
	sqlParams := sql.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "ticketDAO.Delete", zap.String("ID", string(id)))

	_, err := d.dbTrx.GetSqlTx().ExecContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		d.log.Error(ctx, "ticketDAO.Delete", zap.Error(err))
		return err
	}
	return nil
}

func (d ticketDAO) SoftDelete(ctx context.Context, id pubEntity.UUID) error {
	sql := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLUpdate("tickets").
		SetSQLUpdateValue("deleted", true).
		SetSQLWhere("AND", "id", "=", id)

	sqlStr := sql.BuildSQL()
	sqlParams := sql.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "ticketDAO.SoftDelete", zap.String("ID", string(id)))

	_, err := d.dbTrx.GetSqlTx().ExecContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		d.log.Error(ctx, "ticketDAO.SoftDelete", zap.Error(err))
		return err
	}
	return nil
}

func (d ticketDAO) BookStock(ctx context.Context, id pubEntity.UUID, qty int) error {
	if qty <= 0 {
		return fmt.Errorf("invalid qty")
	}

	query := `
        UPDATE tickets
        SET 
            available_qty = available_qty - $1,
            booked_qty    = booked_qty + $1,
            status = CASE 
                WHEN available_qty - $1 <= 0 THEN 'BOOKOUT'::ticket_status_enum
                ELSE 'AVAILABLE'::ticket_status_enum
            END,
            updated_at    = $2
        WHERE id = $3
        AND available_qty >= $1
        AND deleted = false
    `

	d.log.Debug(ctx, "ticketDAO.BookStock", zap.String("ID", string(id)), zap.Int("Qty", qty))

	result, err := d.dbTrx.GetSqlTx().ExecContext(ctx, query, qty, time.Now(), id)
	if err != nil {
		d.log.Error(ctx, "ticketDAO.BookStock", zap.Error(err))
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil || rows == 0 {
		d.log.Warn(ctx, "ticketDAO.BookStock.NoRowsAffected", zap.String("ID", string(id)))
		return fmt.Errorf("insufficient available stock")
	}

	return nil
}

func (d ticketDAO) ConfirmSold(ctx context.Context, id pubEntity.UUID, qty int) error {
	if qty <= 0 {
		return fmt.Errorf("invalid qty")
	}

	query := `
        UPDATE tickets
        SET 
            booked_qty = booked_qty - $1,
            sold_qty   = sold_qty + $1,
            status = CASE 
                WHEN available_qty = 0 AND (booked_qty - $1) <= 0 THEN 'SOLD'::ticket_status_enum
                WHEN available_qty = 0 AND (booked_qty - $1) > 0 THEN 'BOOKOUT'::ticket_status_enum
                ELSE 'AVAILABLE'::ticket_status_enum
            END,
            updated_at = $2
        WHERE id = $3
        AND booked_qty >= $1
        AND deleted = false
    `

	d.log.Debug(ctx, "ticketDAO.ConfirmSold", zap.String("ID", string(id)), zap.Int("Qty", qty))

	result, err := d.dbTrx.GetSqlTx().ExecContext(ctx, query, qty, time.Now(), id)
	if err != nil {
		d.log.Error(ctx, "ticketDAO.ConfirmSold", zap.Error(err))
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil || rows == 0 {
		d.log.Warn(ctx, "ticketDAO.ConfirmSold.NoRowsAffected", zap.String("ID", string(id)))
		return fmt.Errorf("insufficient booked stock")
	}

	return nil
}

func (d ticketDAO) ReleaseBooked(ctx context.Context, id pubEntity.UUID, qty int) error {
	if qty <= 0 {
		return fmt.Errorf("invalid qty")
	}

	query := `
        UPDATE tickets
        SET 
            booked_qty    = booked_qty - $1,
            available_qty = available_qty + $1,
            status = 'AVAILABLE'::ticket_status_enum,
            updated_at    = $2
        WHERE id = $3
        AND booked_qty >= $1
        AND deleted = false
    `

	d.log.Debug(ctx, "ticketDAO.ReleaseBooked", zap.String("ID", string(id)), zap.Int("Qty", qty))

	result, err := d.dbTrx.GetSqlTx().ExecContext(ctx, query, qty, time.Now(), id)
	if err != nil {
		d.log.Error(ctx, "ticketDAO.ReleaseBooked", zap.Error(err))
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil || rows == 0 {
		d.log.Warn(ctx, "ticketDAO.ReleaseBooked.NoRowsAffected", zap.String("ID", string(id)))
		return fmt.Errorf("insufficient booked stock to release")
	}

	return nil
}
