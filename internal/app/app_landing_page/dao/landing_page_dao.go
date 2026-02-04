package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	pubEntity "rakit-tiket-be/pkg/entity"
	entity "rakit-tiket-be/pkg/entity/app_landing_page"

	"gitlab.com/threetopia/sqlgo/v2"
)

type LandingPageDAO interface {
	Search(ctx context.Context, query entity.LandingPageQuery) (entity.LandingPages, error)
	Insert(ctx context.Context, pages entity.LandingPages) error
	Update(ctx context.Context, pages entity.LandingPages) error
	Delete(ctx context.Context, id pubEntity.UUID) error
	SoftDelete(ctx context.Context, id pubEntity.UUID) error
}

type landingPageDAO struct {
	dbTrx DBTransaction
}

func MakeLandingPageDAO(dbTrx DBTransaction) LandingPageDAO {
	return landingPageDAO{
		dbTrx: dbTrx,
	}
}

func (d landingPageDAO) Search(ctx context.Context, query entity.LandingPageQuery) (entity.LandingPages, error) {

	sqlSelect := sqlgo.NewSQLGoSelect().
		SetSQLSelect("lp.id", "id").
		SetSQLSelect("lp.banner_image", "banner_image").
		SetSQLSelect("lp.venue_image", "venue_image").
		SetSQLSelect("lp.event_creator", "event_creator").
		SetSQLSelect("lp.event_name", "event_name").
		SetSQLSelect("lp.event_date", "event_date").
		SetSQLSelect("lp.event_time_start", "event_time_start").
		SetSQLSelect("lp.event_time_end", "event_time_end").
		SetSQLSelect("lp.event_location", "event_location").
		SetSQLSelect("lp.terms_and_conditions", "terms_and_conditions").
		SetSQLSelect("lp.deleted", "deleted").
		SetSQLSelect("lp.data_hash", "data_hash").
		SetSQLSelect("lp.created_at", "created_at").
		SetSQLSelect("lp.updated_at", "updated_at")

	sqlFrom := sqlgo.NewSQLGoFrom().
		SetSQLFrom(`"landing_page_configs"`, "lp")

	sqlWhere := sqlgo.NewSQLGoWhere()

	if len(query.IDs) > 0 {
		sqlWhere.SetSQLWhere("AND", "lp.id", "IN", query.IDs)
	}

	if len(query.EventName) > 0 {
		sqlWhere.SetSQLWhere("AND", "lp.event_name", "IN", query.EventName)
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

	var pages entity.LandingPages
	for rows.Next() {
		var page entity.LandingPage
		var termsJSON []byte

		if err := rows.Scan(
			&page.ID,
			&page.BannerImage,
			&page.VenueImage,
			&page.EventCreator,
			&page.EventName,
			&page.EventDate,
			&page.EventTimeStart,
			&page.EventTimeEnd,
			&page.EventLocation,
			&termsJSON,
			&page.DaoEntity.Deleted,
			&page.DaoEntity.DataHash,
			&page.DaoEntity.CreatedAt,
			&page.DaoEntity.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if len(termsJSON) > 0 {
			if err := json.Unmarshal(termsJSON, &page.TermsAndConditions); err != nil {
				page.TermsAndConditions = []string{}
			}
		}

		pages = append(pages, page)
	}

	return pages, nil
}

func (d landingPageDAO) Insert(ctx context.Context, pages entity.LandingPages) error {

	if len(pages) < 1 {
		return fmt.Errorf("empty landing page data")
	}

	sqlInsert := sqlgo.NewSQLGoInsert().
		SetSQLInsert("landing_page_configs").
		SetSQLInsertColumn(
			"id",
			"banner_image",
			"venue_image",
			"event_creator",
			"event_name",
			"event_date",
			"event_time_start",
			"event_time_end",
			"event_location",
			"terms_and_conditions",
			"data_hash",
			"created_at",
		)

	for i, page := range pages {
		page.CreatedAt = time.Now()
		page.ID = pubEntity.MakeUUID(
			page.DataHash.String(),
			page.CreatedAt.String(),
		)

		termsJSON, err := json.Marshal(page.TermsAndConditions)
		if err != nil {
			termsJSON = []byte("[]")
		}

		sqlInsert.SetSQLInsertValue(
			page.ID,
			page.BannerImage,
			page.VenueImage,
			page.EventCreator,
			page.EventName,
			page.EventDate,
			page.EventTimeStart,
			page.EventTimeEnd,
			page.EventLocation,
			termsJSON,
			page.DataHash,
			page.CreatedAt,
		)

		pages[i] = page
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

func (d landingPageDAO) Update(ctx context.Context, pages entity.LandingPages) error {

	if len(pages) < 1 {
		return fmt.Errorf("empty landing page data")
	}

	for i, page := range pages {
		now := time.Now()
		page.UpdatedAt = &now

		termsJSON, err := json.Marshal(page.TermsAndConditions)
		if err != nil {
			// Jika gagal, set array kosong default JSON
			termsJSON = []byte("[]")
		}

		sql := sqlgo.NewSQLGo().
			SetSQLSchema("public").
			SetSQLUpdate("landing_page_configs").
			SetSQLUpdateValue("banner_image", page.BannerImage).
			SetSQLUpdateValue("venue_image", page.VenueImage).
			SetSQLUpdateValue("event_creator", page.EventCreator).
			SetSQLUpdateValue("event_name", page.EventName).
			SetSQLUpdateValue("event_date", page.EventDate).
			SetSQLUpdateValue("event_time_start", page.EventTimeStart).
			SetSQLUpdateValue("event_time_end", page.EventTimeEnd).
			SetSQLUpdateValue("event_location", page.EventLocation).
			SetSQLUpdateValue("terms_and_conditions", termsJSON).
			SetSQLUpdateValue("data_hash", page.DataHash).
			SetSQLUpdateValue("updated_at", page.UpdatedAt).
			SetSQLWhere("AND", "id", "=", page.ID)

		_, err = d.dbTrx.GetSqlTx().ExecContext(
			ctx,
			sql.BuildSQL(),
			sql.GetSQLGoParameter().GetSQLParameter()...,
		)
		if err != nil {
			return err
		}

		pages[i] = page
	}

	return nil
}

func (d landingPageDAO) Delete(ctx context.Context, id pubEntity.UUID) error {
	sql := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLDelete("landing_page_configs").
		SetSQLWhere("AND", "id", "=", id)

	_, err := d.dbTrx.GetSqlTx().ExecContext(
		ctx,
		sql.BuildSQL(),
		sql.GetSQLGoParameter().GetSQLParameter()...,
	)

	return err
}

func (d landingPageDAO) SoftDelete(ctx context.Context, id pubEntity.UUID) error {
	sql := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLUpdate("landing_page_configs").
		SetSQLUpdateValue("deleted", true).
		SetSQLWhere("AND", "id", "=", id)

	_, err := d.dbTrx.GetSqlTx().ExecContext(
		ctx,
		sql.BuildSQL(),
		sql.GetSQLGoParameter().GetSQLParameter()...,
	)

	return err
}
