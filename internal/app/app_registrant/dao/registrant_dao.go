package dao

import (
	"context"
	"fmt"
	"time"

	pubEntity "rakit-tiket-be/pkg/entity"
	entity "rakit-tiket-be/pkg/entity/app_registrant"

	"gitlab.com/threetopia/sqlgo/v2"
)

type RegistrantDAO interface {
	Search(ctx context.Context, query entity.RegistrantQuery) (entity.Registrants, error)
	Insert(ctx context.Context, registrants entity.Registrants) error
	Update(ctx context.Context, registrants entity.Registrants) error
	Delete(ctx context.Context, id pubEntity.UUID) error
	SoftDelete(ctx context.Context, id pubEntity.UUID) error
}

func nullStr(s *string) interface{} {
	if s == nil {
		return nil
	}
	return *s
}

func nullTime(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return *t
}

func nullUUID(u *pubEntity.UUID) interface{} {
	if u == nil {
		return nil
	}
	return *u
}

type registrantDAO struct {
	dbTrx DBTransaction
}

func MakeRegistrantDAO(dbTrx DBTransaction) RegistrantDAO {
	return registrantDAO{
		dbTrx: dbTrx,
	}
}

func (d registrantDAO) Search(ctx context.Context, query entity.RegistrantQuery) (entity.Registrants, error) {

	sqlSelect := sqlgo.NewSQLGoSelect().
		SetSQLSelect("r.id", "id").
		SetSQLSelect("r.unique_code", "unique_code").
		SetSQLSelect("r.ticket_id", "ticket_id").
		SetSQLSelect("r.name", "name").
		SetSQLSelect("r.email", "email").
		SetSQLSelect("r.phone", "phone").
		SetSQLSelect("r.gender", "gender").
		SetSQLSelect("r.birthdate", "birthdate").
		SetSQLSelect("r.total_cost", "total_cost").
		SetSQLSelect("r.total_tickets", "total_tickets").
		SetSQLSelect("r.status", "status").
		SetSQLSelect("r.created_at", "created_at").
		SetSQLSelect("r.updated_at", "updated_at")

	sqlFrom := sqlgo.NewSQLGoFrom().
		SetSQLFrom("registrants", "r")

	sqlWhere := sqlgo.NewSQLGoWhere()

	// Default filter: jangan ambil yang sudah soft deleted
	sqlWhere.SetSQLWhere("AND", "r.deleted", "=", false)

	if len(query.IDs) > 0 {
		sqlWhere.SetSQLWhere("AND", "r.id", "IN", query.IDs)
	}

	if len(query.UniqueCodes) > 0 {
		sqlWhere.SetSQLWhere("AND", "r.unique_code", "IN", query.UniqueCodes)
	}

	if len(query.Emails) > 0 {
		sqlWhere.SetSQLWhere("AND", "r.email", "IN", query.Emails)
	}

	if len(query.Statuses) > 0 {
		sqlWhere.SetSQLWhere("AND", "r.status", "IN", query.Statuses)
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

	var registrants entity.Registrants
	for rows.Next() {
		var reg entity.Registrant

		// Note: Scan urutannya harus sama dengan SetSQLSelect
		if err := rows.Scan(
			&reg.ID,
			&reg.UniqueCode,
			nullUUID(reg.TicketID),
			&reg.Name,
			&reg.Email,
			&reg.Phone,
			nullStr(reg.Gender),
			nullTime(reg.Birthdate),
			&reg.TotalCost,
			&reg.TotalTickets,
			&reg.Status,
			&reg.CreatedAt,
			&reg.UpdatedAt,
		); err != nil {
			return nil, err
		}

		registrants = append(registrants, reg)
	}

	return registrants, nil
}

func (d registrantDAO) Insert(ctx context.Context, registrants entity.Registrants) error {

	if len(registrants) < 1 {
		return fmt.Errorf("empty registrant data")
	}

	sqlInsert := sqlgo.NewSQLGoInsert().
		SetSQLInsert("registrants").
		SetSQLInsertColumn(
			"id",
			"unique_code",
			"ticket_id",
			"name",
			"email",
			"phone",
			"gender",
			"birthdate",
			"total_cost",
			"total_tickets",
			"status",
			"data_hash",
			"created_at",
		)

	for i, reg := range registrants {
		reg.CreatedAt = time.Now()

		if reg.ID == "" {
			reg.ID = pubEntity.MakeUUID(
				reg.UniqueCode,
				reg.Email,
				reg.CreatedAt.String(),
			)
		}

		sqlInsert.SetSQLInsertValue(
			reg.ID,
			reg.UniqueCode,
			reg.TicketID,
			reg.Name,
			reg.Email,
			reg.Phone,
			reg.Gender,
			reg.Birthdate,
			reg.TotalCost,
			reg.TotalTickets,
			reg.Status,
			reg.DaoEntity.DataHash,
			reg.DaoEntity.CreatedAt,
		)

		registrants[i] = reg
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

func (d registrantDAO) Update(ctx context.Context, registrants entity.Registrants) error {

	if len(registrants) < 1 {
		return fmt.Errorf("empty registrant data")
	}

	for i, reg := range registrants {
		now := time.Now()
		reg.UpdatedAt = &now

		sql := sqlgo.NewSQLGo().
			SetSQLSchema("public").
			SetSQLUpdate("registrants").
			SetSQLUpdateValue("ticket_id", reg.TicketID).
			SetSQLUpdateValue("name", reg.Name).
			SetSQLUpdateValue("email", reg.Email).
			SetSQLUpdateValue("phone", reg.Phone).
			SetSQLUpdateValue("gender", reg.Gender).
			SetSQLUpdateValue("birthdate", reg.Birthdate).
			SetSQLUpdateValue("total_cost", reg.TotalCost).
			SetSQLUpdateValue("total_tickets", reg.TotalTickets).
			SetSQLUpdateValue("status", reg.Status).
			SetSQLUpdateValue("updated_at", reg.UpdatedAt).
			SetSQLWhere("AND", "id", "=", reg.ID)

		_, err := d.dbTrx.GetSqlTx().ExecContext(
			ctx,
			sql.BuildSQL(),
			sql.GetSQLGoParameter().GetSQLParameter()...,
		)
		if err != nil {
			return err
		}

		registrants[i] = reg
	}

	return nil
}

func (d registrantDAO) Delete(ctx context.Context, id pubEntity.UUID) error {
	sql := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLDelete("registrants").
		SetSQLWhere("AND", "id", "=", id)

	_, err := d.dbTrx.GetSqlTx().ExecContext(
		ctx,
		sql.BuildSQL(),
		sql.GetSQLGoParameter().GetSQLParameter()...,
	)

	return err
}

func (d registrantDAO) SoftDelete(ctx context.Context, id pubEntity.UUID) error {
	sql := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLUpdate("registrants").
		SetSQLUpdateValue("deleted", true).
		SetSQLWhere("AND", "id", "=", id)

	_, err := d.dbTrx.GetSqlTx().ExecContext(
		ctx,
		sql.BuildSQL(),
		sql.GetSQLGoParameter().GetSQLParameter()...,
	)

	return err
}
