package dao

import (
	"context"
	"fmt"
	"time"

	pubEntity "rakit-tiket-be/pkg/entity"
	entity "rakit-tiket-be/pkg/entity/app_event" // Sesuaikan path package entity event Anda
	"rakit-tiket-be/pkg/util"

	"gitlab.com/threetopia/sqlgo/v2"
	"go.uber.org/zap"
)

type EventDAO interface {
	Search(ctx context.Context, query entity.EventQuery) (entity.Events, error)
	Insert(ctx context.Context, events entity.Events) error
	Update(ctx context.Context, events entity.Events) error
	Delete(ctx context.Context, id pubEntity.UUID) error
	SoftDelete(ctx context.Context, id pubEntity.UUID) error
}

type eventDAO struct {
	log   util.LogUtil
	dbTrx DBTransaction
}

func MakeEventDAO(log util.LogUtil, dbTrx DBTransaction) EventDAO {
	return eventDAO{
		log:   log,
		dbTrx: dbTrx,
	}
}

func (d eventDAO) Search(ctx context.Context, query entity.EventQuery) (entity.Events, error) {
	sqlSelect := sqlgo.NewSQLGoSelect().
		SetSQLSelect("e.id", "id").
		SetSQLSelect("e.slug", "slug").
		SetSQLSelect("e.name", "name").
		SetSQLSelect("e.status", "status").
		SetSQLSelect("e.ticket_prefix_code", "ticket_prefix_code").
		SetSQLSelect("e.max_ticket_per_tx", "max_ticket_per_tx").
		SetSQLSelect("e.deleted", "deleted").
		SetSQLSelect("e.data_hash", "data_hash").
		SetSQLSelect("e.created_at", "created_at").
		SetSQLSelect("e.updated_at", "updated_at")

	sqlFrom := sqlgo.NewSQLGoFrom().
		SetSQLFrom("events", "e")

	sqlWhere := sqlgo.NewSQLGoWhere()
	// Selalu filter yang belum dihapus kecuali ditentukan lain
	sqlWhere.SetSQLWhere("AND", "e.deleted", "=", false)

	if len(query.IDs) > 0 {
		sqlWhere.SetSQLWhere("AND", "e.id", "IN", query.IDs)
	}
	if len(query.Slugs) > 0 {
		sqlWhere.SetSQLWhere("AND", "e.slug", "IN", query.Slugs)
	}
	if len(query.Statuses) > 0 {
		sqlWhere.SetSQLWhere("AND", "e.status", "IN", query.Statuses)
	}

	sql := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLGoSelect(sqlSelect).
		SetSQLGoFrom(sqlFrom).
		SetSQLGoWhere(sqlWhere)

	sqlStr := sql.BuildSQL()
	sqlParams := sql.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "eventDAO.Search",
		zap.String("SQL", sqlStr),
		zap.Any("Params", sqlParams),
	)

	rows, err := d.dbTrx.GetSqlDB().QueryContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		d.log.Error(ctx, "eventDAO.Search", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var events entity.Events
	for rows.Next() {
		var event entity.Event
		if err := rows.Scan(
			&event.ID, &event.Slug, &event.Name, &event.Status,
			&event.TicketPrefixCode, &event.MaxTicketPerTx,
			&event.Deleted, &event.DataHash,
			&event.CreatedAt, &event.UpdatedAt,
		); err != nil {
			d.log.Error(ctx, "eventDAO.Search.Scan", zap.Error(err))
			return nil, err
		}
		events = append(events, event)
	}

	return events, nil
}

func (d eventDAO) Insert(ctx context.Context, events entity.Events) error {
	if len(events) < 1 {
		return fmt.Errorf("empty event data")
	}

	sqlInsert := sqlgo.NewSQLGoInsert().
		SetSQLInsert("events").
		SetSQLInsertColumn(
			"id", "slug", "name", "status", "ticket_prefix_code",
			"max_ticket_per_tx", "deleted", "data_hash", "created_at",
		)

	for i, event := range events {
		event.CreatedAt = time.Now()
		if event.ID == "" {
			// Menggunakan Name sebagai salt seed untuk UUID baru
			event.ID = pubEntity.MakeUUID(event.Name, event.CreatedAt.String())
		}

		sqlInsert.SetSQLInsertValue(
			event.ID, event.Slug, event.Name, event.Status, event.TicketPrefixCode,
			event.MaxTicketPerTx, event.Deleted, event.DataHash, event.CreatedAt,
		)
		events[i] = event
	}

	sql := sqlgo.NewSQLGo().SetSQLSchema("public").SetSQLGoInsert(sqlInsert)
	sqlStr := sql.BuildSQL()
	sqlParams := sql.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "eventDAO.Insert", zap.String("SQL", sqlStr), zap.Int("Count", len(events)))

	_, err := d.dbTrx.GetSqlTx().ExecContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		d.log.Error(ctx, "eventDAO.Insert", zap.Error(err))
		return err
	}
	return nil
}

func (d eventDAO) Update(ctx context.Context, events entity.Events) error {
	if len(events) < 1 {
		return fmt.Errorf("empty event data")
	}

	for i, event := range events {
		now := time.Now()
		event.UpdatedAt = &now

		sql := sqlgo.NewSQLGo().
			SetSQLSchema("public").
			SetSQLUpdate("events").
			SetSQLUpdateValue("slug", event.Slug).
			SetSQLUpdateValue("name", event.Name).
			SetSQLUpdateValue("status", event.Status).
			SetSQLUpdateValue("ticket_prefix_code", event.TicketPrefixCode).
			SetSQLUpdateValue("max_ticket_per_tx", event.MaxTicketPerTx).
			SetSQLUpdateValue("data_hash", event.DataHash).
			SetSQLUpdateValue("updated_at", event.UpdatedAt).
			SetSQLWhere("AND", "id", "=", event.ID)

		sqlStr := sql.BuildSQL()
		sqlParams := sql.GetSQLGoParameter().GetSQLParameter()

		d.log.Debug(ctx, "eventDAO.Update", zap.String("ID", string(event.ID)))

		_, err := d.dbTrx.GetSqlTx().ExecContext(ctx, sqlStr, sqlParams...)
		if err != nil {
			d.log.Error(ctx, "eventDAO.Update", zap.Error(err))
			return err
		}
		events[i] = event
	}
	return nil
}

func (d eventDAO) Delete(ctx context.Context, id pubEntity.UUID) error {
	sql := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLDelete("events").
		SetSQLWhere("AND", "id", "=", id)

	sqlStr := sql.BuildSQL()
	sqlParams := sql.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "eventDAO.Delete", zap.String("ID", string(id)))

	_, err := d.dbTrx.GetSqlTx().ExecContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		d.log.Error(ctx, "eventDAO.Delete", zap.Error(err))
		return err
	}
	return nil
}

func (d eventDAO) SoftDelete(ctx context.Context, id pubEntity.UUID) error {
	now := time.Now()
	sql := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLUpdate("events").
		SetSQLUpdateValue("deleted", true).
		SetSQLUpdateValue("updated_at", now).
		SetSQLWhere("AND", "id", "=", id)

	sqlStr := sql.BuildSQL()
	sqlParams := sql.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "eventDAO.SoftDelete", zap.String("ID", string(id)))

	_, err := d.dbTrx.GetSqlTx().ExecContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		d.log.Error(ctx, "eventDAO.SoftDelete", zap.Error(err))
		return err
	}
	return nil
}
