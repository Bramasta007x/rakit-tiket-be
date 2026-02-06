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
		// Event Info
		SetSQLSelect("lp.event_name", "event_name").
		SetSQLSelect("lp.event_subtitle", "event_subtitle").
		SetSQLSelect("lp.event_creator", "event_creator").
		SetSQLSelect("lp.event_date", "event_date").
		SetSQLSelect("lp.event_time_start", "event_time_start").
		SetSQLSelect("lp.event_time_end", "event_time_end").
		SetSQLSelect("lp.event_location", "event_location").
		SetSQLSelect("lp.logo_image", "logo_image").
		// Hero Section
		SetSQLSelect("lp.hero_id", "hero_id").
		SetSQLSelect("lp.banner_image", "banner_image").
		SetSQLSelect("lp.banner_color", "banner_color").
		SetSQLSelect("lp.hero_button_id", "hero_button_id").
		SetSQLSelect("lp.hero_button_text", "hero_button_text").
		SetSQLSelect("lp.hero_button_link", "hero_button_link").
		// Countdown
		SetSQLSelect("lp.hero_countdown_id", "hero_countdown_id").
		SetSQLSelect("lp.hero_countdown_date", "hero_countdown_date").
		SetSQLSelect("lp.hero_countdown_time_start", "hero_countdown_time_start").
		SetSQLSelect("lp.hero_countdown_time_end", "hero_countdown_time_end").
		SetSQLSelect("lp.hero_countdown_after_text", "hero_countdown_after_text").
		// Venue Section
		SetSQLSelect("lp.venue_id", "venue_id").
		SetSQLSelect("lp.venue_image", "venue_image").
		SetSQLSelect("lp.venue_layout", "venue_layout").
		SetSQLSelect("lp.venue_address", "venue_address").
		SetSQLSelect("lp.venue_map_link", "venue_map_link").
		// JSON Data
		SetSQLSelect("lp.terms_and_conditions", "terms_and_conditions").
		SetSQLSelect("lp.faqs", "faqs").
		// Metadata
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
		var faqsJSON []byte

		if err := rows.Scan(
			&page.ID,
			// Event Info
			&page.EventName,
			&page.EventSubtitle,
			&page.EventCreator,
			&page.EventDate,
			&page.EventTimeStart,
			&page.EventTimeEnd,
			&page.EventLocation,
			&page.LogoImage,
			// Hero Section
			&page.HeroID,
			&page.BannerImage,
			&page.BannerColor,
			&page.HeroButtonID,
			&page.HeroButtonText,
			&page.HeroButtonLink,
			// Countdown
			&page.HeroCountdownID,
			&page.HeroCountdownDate,
			&page.HeroCountdownTimeStart,
			&page.HeroCountdownTimeEnd,
			&page.HeroCountdownAfterText,
			// Venue Section
			&page.VenueID,
			&page.VenueImage,
			&page.VenueLayout,
			&page.VenueAddress,
			&page.VenueMapLink,
			// JSON Data
			&termsJSON,
			&faqsJSON,
			// Metadata
			&page.DaoEntity.Deleted,
			&page.DaoEntity.DataHash,
			&page.DaoEntity.CreatedAt,
			&page.DaoEntity.UpdatedAt,
		); err != nil {
			return nil, err
		}

		// Unmarshal Terms
		if len(termsJSON) > 0 {
			if err := json.Unmarshal(termsJSON, &page.TermsAndConditions); err != nil {
				page.TermsAndConditions = []string{}
			}
		}

		// Unmarshal FAQs
		if len(faqsJSON) > 0 {
			if err := json.Unmarshal(faqsJSON, &page.Faqs); err != nil {
				page.Faqs = []string{}
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
			// Event
			"event_name",
			"event_subtitle",
			"event_creator",
			"event_date",
			"event_time_start",
			"event_time_end",
			"event_location",
			"logo_image",
			// Hero
			"hero_id",
			"banner_image",
			"banner_color",
			"hero_button_id",
			"hero_button_text",
			"hero_button_link",
			// Countdown
			"hero_countdown_id",
			"hero_countdown_date",
			"hero_countdown_time_start",
			"hero_countdown_time_end",
			"hero_countdown_after_text",
			// Venue
			"venue_id",
			"venue_image",
			"venue_layout",
			"venue_address",
			"venue_map_link",
			// JSON
			"terms_and_conditions",
			"faqs",
			// Metadata
			"data_hash",
			"created_at",
		)

	for i, page := range pages {
		page.CreatedAt = time.Now()
		page.ID = pubEntity.MakeUUID(
			page.DataHash.String(),
			page.CreatedAt.String(),
		)

		// Marshal Terms
		termsJSON, err := json.Marshal(page.TermsAndConditions)
		if err != nil {
			termsJSON = []byte("[]")
		}

		// Marshal FAQs
		faqsJSON, err := json.Marshal(page.Faqs)
		if err != nil {
			faqsJSON = []byte("[]")
		}

		sqlInsert.SetSQLInsertValue(
			page.ID,
			// Event
			page.EventName,
			page.EventSubtitle,
			page.EventCreator,
			page.EventDate,
			page.EventTimeStart,
			page.EventTimeEnd,
			page.EventLocation,
			page.LogoImage,
			// Hero
			page.HeroID,
			page.BannerImage,
			page.BannerColor,
			page.HeroButtonID,
			page.HeroButtonText,
			page.HeroButtonLink,
			// Countdown
			page.HeroCountdownID,
			page.HeroCountdownDate,
			page.HeroCountdownTimeStart,
			page.HeroCountdownTimeEnd,
			page.HeroCountdownAfterText,
			// Venue
			page.VenueID,
			page.VenueImage,
			page.VenueLayout,
			page.VenueAddress,
			page.VenueMapLink,
			// JSON
			termsJSON,
			faqsJSON,
			// Metadata
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
			termsJSON = []byte("[]")
		}

		faqsJSON, err := json.Marshal(page.Faqs)
		if err != nil {
			faqsJSON = []byte("[]")
		}

		sql := sqlgo.NewSQLGo().
			SetSQLSchema("public").
			SetSQLUpdate("landing_page_configs").
			// Event
			SetSQLUpdateValue("event_name", page.EventName).
			SetSQLUpdateValue("event_subtitle", page.EventSubtitle).
			SetSQLUpdateValue("event_creator", page.EventCreator).
			SetSQLUpdateValue("event_date", page.EventDate).
			SetSQLUpdateValue("event_time_start", page.EventTimeStart).
			SetSQLUpdateValue("event_time_end", page.EventTimeEnd).
			SetSQLUpdateValue("event_location", page.EventLocation).
			SetSQLUpdateValue("logo_image", page.LogoImage).
			// Hero
			SetSQLUpdateValue("hero_id", page.HeroID).
			SetSQLUpdateValue("banner_image", page.BannerImage).
			SetSQLUpdateValue("banner_color", page.BannerColor).
			SetSQLUpdateValue("hero_button_id", page.HeroButtonID).
			SetSQLUpdateValue("hero_button_text", page.HeroButtonText).
			SetSQLUpdateValue("hero_button_link", page.HeroButtonLink).
			// Countdown
			SetSQLUpdateValue("hero_countdown_id", page.HeroCountdownID).
			SetSQLUpdateValue("hero_countdown_date", page.HeroCountdownDate).
			SetSQLUpdateValue("hero_countdown_time_start", page.HeroCountdownTimeStart).
			SetSQLUpdateValue("hero_countdown_time_end", page.HeroCountdownTimeEnd).
			SetSQLUpdateValue("hero_countdown_after_text", page.HeroCountdownAfterText).
			// Venue
			SetSQLUpdateValue("venue_id", page.VenueID).
			SetSQLUpdateValue("venue_image", page.VenueImage).
			SetSQLUpdateValue("venue_layout", page.VenueLayout).
			SetSQLUpdateValue("venue_address", page.VenueAddress).
			SetSQLUpdateValue("venue_map_link", page.VenueMapLink).
			// JSON
			SetSQLUpdateValue("terms_and_conditions", termsJSON).
			SetSQLUpdateValue("faqs", faqsJSON).
			// Metadata
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
