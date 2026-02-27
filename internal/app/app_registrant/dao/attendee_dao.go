package dao

import (
	"context"
	"fmt"
	"time"

	pubEntity "rakit-tiket-be/pkg/entity"
	entity "rakit-tiket-be/pkg/entity/app_registrant"

	"gitlab.com/threetopia/sqlgo/v2"
)

type AttendeeDAO interface {
	Search(ctx context.Context, query entity.AttendeeQuery) (entity.Attendees, error)
	Insert(ctx context.Context, attendees entity.Attendees) error
	Update(ctx context.Context, attendees entity.Attendees) error
	Delete(ctx context.Context, id pubEntity.UUID) error
	SoftDelete(ctx context.Context, id pubEntity.UUID) error
}

func attNullStr(s *string) interface{} {
	if s == nil {
		return nil
	}
	return *s
}

func attNullTime(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return *t
}

type attendeeDAO struct {
	dbTrx DBTransaction
}

func MakeAttendeeDAO(dbTrx DBTransaction) AttendeeDAO {
	return attendeeDAO{
		dbTrx: dbTrx,
	}
}

func (d attendeeDAO) Search(ctx context.Context, query entity.AttendeeQuery) (entity.Attendees, error) {

	sqlSelect := sqlgo.NewSQLGoSelect().
		SetSQLSelect("a.id", "id").
		SetSQLSelect("a.registrant_id", "registrant_id").
		SetSQLSelect("a.ticket_id", "ticket_id").
		SetSQLSelect("a.name", "name").
		SetSQLSelect("a.gender", "gender").
		SetSQLSelect("a.birthdate", "birthdate").
		SetSQLSelect("a.created_at", "created_at").
		SetSQLSelect("a.updated_at", "updated_at")

	sqlFrom := sqlgo.NewSQLGoFrom().
		SetSQLFrom("attendees", "a")

	sqlWhere := sqlgo.NewSQLGoWhere()
	sqlWhere.SetSQLWhere("AND", "a.deleted", "=", false)

	if len(query.IDs) > 0 {
		sqlWhere.SetSQLWhere("AND", "a.id", "IN", query.IDs)
	}

	if len(query.RegistrantIDs) > 0 {
		sqlWhere.SetSQLWhere("AND", "a.registrant_id", "IN", query.RegistrantIDs)
	}

	if len(query.TicketIDs) > 0 {
		sqlWhere.SetSQLWhere("AND", "a.ticket_id", "IN", query.TicketIDs)
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

	var attendees entity.Attendees
	for rows.Next() {
		var att entity.Attendee

		if err := rows.Scan(
			&att.ID,
			&att.RegistrantID,
			&att.TicketID,
			&att.Name,
			&att.Gender,
			&att.Birthdate,
			&att.CreatedAt,
			&att.UpdatedAt,
		); err != nil {
			return nil, err
		}

		attendees = append(attendees, att)
	}

	return attendees, nil
}

func (d attendeeDAO) Insert(ctx context.Context, attendees entity.Attendees) error {

	if len(attendees) < 1 {
		return fmt.Errorf("empty attendee data")
	}

	sqlInsert := sqlgo.NewSQLGoInsert().
		SetSQLInsert("attendees").
		SetSQLInsertColumn(
			"id",
			"registrant_id",
			"ticket_id",
			"name",
			"gender",
			"birthdate",
			"data_hash",
			"created_at",
		)

	for i, att := range attendees {
		att.CreatedAt = time.Now()

		if att.ID == "" {
			att.ID = pubEntity.MakeUUID(
				att.Name,
				string(att.RegistrantID),
				att.CreatedAt.String(),
			)
		}

		sqlInsert.SetSQLInsertValue(
			att.ID,
			att.RegistrantID,
			att.TicketID,
			att.Name,
			attNullStr(att.Gender),
			attNullTime(att.Birthdate),
			att.DaoEntity.DataHash,
			att.DaoEntity.CreatedAt,
		)

		attendees[i] = att
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

func (d attendeeDAO) Update(ctx context.Context, attendees entity.Attendees) error {

	if len(attendees) < 1 {
		return fmt.Errorf("empty attendee data")
	}

	for i, att := range attendees {
		now := time.Now()
		att.UpdatedAt = &now

		sql := sqlgo.NewSQLGo().
			SetSQLSchema("public").
			SetSQLUpdate("attendees").
			SetSQLUpdateValue("name", att.Name).
			SetSQLUpdateValue("gender", att.Gender).
			SetSQLUpdateValue("birthdate", att.Birthdate).
			SetSQLUpdateValue("ticket_id", att.TicketID).
			SetSQLUpdateValue("updated_at", att.UpdatedAt).
			SetSQLWhere("AND", "id", "=", att.ID)

		_, err := d.dbTrx.GetSqlTx().ExecContext(
			ctx,
			sql.BuildSQL(),
			sql.GetSQLGoParameter().GetSQLParameter()...,
		)
		if err != nil {
			return err
		}

		attendees[i] = att
	}

	return nil
}

func (d attendeeDAO) Delete(ctx context.Context, id pubEntity.UUID) error {
	sql := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLDelete("attendees").
		SetSQLWhere("AND", "id", "=", id)

	_, err := d.dbTrx.GetSqlTx().ExecContext(
		ctx,
		sql.BuildSQL(),
		sql.GetSQLGoParameter().GetSQLParameter()...,
	)

	return err
}

func (d attendeeDAO) SoftDelete(ctx context.Context, id pubEntity.UUID) error {
	sql := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLUpdate("attendees").
		SetSQLUpdateValue("deleted", true).
		SetSQLWhere("AND", "id", "=", id)

	_, err := d.dbTrx.GetSqlTx().ExecContext(
		ctx,
		sql.BuildSQL(),
		sql.GetSQLGoParameter().GetSQLParameter()...,
	)

	return err
}
