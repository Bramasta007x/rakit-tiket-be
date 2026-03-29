package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	pubEntity "rakit-tiket-be/pkg/entity"
	entity "rakit-tiket-be/pkg/entity/app_artist"
	"rakit-tiket-be/pkg/util"

	"gitlab.com/threetopia/sqlgo/v2"
	"go.uber.org/zap"
)

type ArtistDAO interface {
	Search(ctx context.Context, query entity.ArtistQuery) (entity.Artists, error)
	SearchByID(ctx context.Context, id pubEntity.UUID) (entity.Artist, error)
	Insert(ctx context.Context, artists entity.Artists) error
	Update(ctx context.Context, artists entity.Artists) error
	Delete(ctx context.Context, id pubEntity.UUID) error
	SoftDelete(ctx context.Context, id pubEntity.UUID) error
}

type artistDAO struct {
	log   util.LogUtil
	dbTrx DBTransaction
}

func MakeArtistDAO(log util.LogUtil, dbTrx DBTransaction) ArtistDAO {
	return artistDAO{
		log:   log,
		dbTrx: dbTrx,
	}
}

func (d artistDAO) Search(ctx context.Context, query entity.ArtistQuery) (entity.Artists, error) {
	sqlSelect := sqlgo.NewSQLGoSelect().
		SetSQLSelect("a.id", "id").
		SetSQLSelect("a.image", "image").
		SetSQLSelect("a.name", "name").
		SetSQLSelect("a.genre", "genre").
		SetSQLSelect("a.artist_social_media", "artist_social_media").
		SetSQLSelect("a.deleted", "deleted").
		SetSQLSelect("a.data_hash", "data_hash").
		SetSQLSelect("a.created_at", "created_at").
		SetSQLSelect("a.updated_at", "updated_at")

	sqlFrom := sqlgo.NewSQLGoFrom().
		SetSQLFrom("artists", "a")

	sqlWhere := sqlgo.NewSQLGoWhere()
	sqlWhere.SetSQLWhere("AND", "a.deleted", "=", false)

	if len(query.IDs) > 0 {
		sqlWhere.SetSQLWhere("AND", "a.id", "IN", query.IDs)
	}
	if len(query.Names) > 0 {
		sqlWhere.SetSQLWhere("AND", "a.name", "IN", query.Names)
	}
	if len(query.Genres) > 0 {
		sqlWhere.SetSQLWhere("AND", "a.genre", "IN", query.Genres)
	}

	sql := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLGoSelect(sqlSelect).
		SetSQLGoFrom(sqlFrom).
		SetSQLGoWhere(sqlWhere)

	sqlStr := sql.BuildSQL()
	sqlParams := sql.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "artistDAO.Search",
		zap.String("SQL", sqlStr),
		zap.Any("Params", sqlParams),
	)

	rows, err := d.dbTrx.GetSqlDB().QueryContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		d.log.Error(ctx, "artistDAO.Search", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var artists entity.Artists
	for rows.Next() {
		var artist entity.Artist
		var socialMediaJSON []byte

		if err := rows.Scan(
			&artist.ID, &artist.Image, &artist.Name, &artist.Genre,
			&socialMediaJSON,
			&artist.Deleted, &artist.DataHash,
			&artist.CreatedAt, &artist.UpdatedAt,
		); err != nil {
			d.log.Error(ctx, "artistDAO.Search.Scan", zap.Error(err))
			return nil, err
		}

		artist.ImageUrl = artist.Image

		if len(socialMediaJSON) > 0 {
			if err := json.Unmarshal(socialMediaJSON, &artist.ArtistSocialMedia); err != nil {
				artist.ArtistSocialMedia = []entity.ArtistSocialMedia{}
			}
		}

		artists = append(artists, artist)
	}

	return artists, nil
}

func (d artistDAO) SearchByID(ctx context.Context, id pubEntity.UUID) (entity.Artist, error) {
	sqlSelect := sqlgo.NewSQLGoSelect().
		SetSQLSelect("a.id", "id").
		SetSQLSelect("a.image", "image").
		SetSQLSelect("a.name", "name").
		SetSQLSelect("a.genre", "genre").
		SetSQLSelect("a.role", "role").
		SetSQLSelect("a.artist_social_media", "artist_social_media").
		SetSQLSelect("a.deleted", "deleted").
		SetSQLSelect("a.data_hash", "data_hash").
		SetSQLSelect("a.created_at", "created_at").
		SetSQLSelect("a.updated_at", "updated_at")

	sqlFrom := sqlgo.NewSQLGoFrom().
		SetSQLFrom("artists", "a")

	sqlWhere := sqlgo.NewSQLGoWhere()
	sqlWhere.SetSQLWhere("AND", "a.deleted", "=", false)
	sqlWhere.SetSQLWhere("AND", "a.id", "=", id)

	sql := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLGoSelect(sqlSelect).
		SetSQLGoFrom(sqlFrom).
		SetSQLGoWhere(sqlWhere)

	sqlStr := sql.BuildSQL()
	sqlParams := sql.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "artistDAO.SearchByID",
		zap.String("SQL", sqlStr),
		zap.Any("Params", sqlParams),
	)

	row := d.dbTrx.GetSqlDB().QueryRowContext(ctx, sqlStr, sqlParams...)

	var artist entity.Artist
	var socialMediaJSON []byte

	if err := row.Scan(
		&artist.ID, &artist.Image, &artist.Name, &artist.Genre,
		&socialMediaJSON,
		&artist.Deleted, &artist.DataHash,
		&artist.CreatedAt, &artist.UpdatedAt,
	); err != nil {
		d.log.Error(ctx, "artistDAO.SearchByID.Scan", zap.Error(err))
		return entity.Artist{}, err
	}

	artist.ImageUrl = artist.Image

	if len(socialMediaJSON) > 0 {
		if err := json.Unmarshal(socialMediaJSON, &artist.ArtistSocialMedia); err != nil {
			artist.ArtistSocialMedia = []entity.ArtistSocialMedia{}
		}
	}

	return artist, nil
}

func (d artistDAO) Insert(ctx context.Context, artists entity.Artists) error {
	if len(artists) < 1 {
		return fmt.Errorf("empty artist data")
	}

	sqlInsert := sqlgo.NewSQLGoInsert().
		SetSQLInsert("artists").
		SetSQLInsertColumn(
			"id", "image", "name", "genre", "artist_social_media",
			"deleted", "data_hash", "created_at",
		)

	for i, artist := range artists {
		artist.CreatedAt = time.Now()
		if artist.ID == "" {
			artist.ID = pubEntity.MakeUUID(artist.Name, artist.CreatedAt.String())
		}

		socialMediaJSON, err := json.Marshal(artist.ArtistSocialMedia)
		if err != nil {
			socialMediaJSON = []byte("[]")
		}

		sqlInsert.SetSQLInsertValue(
			artist.ID, artist.Image, artist.Name, artist.Genre,
			socialMediaJSON,
			artist.Deleted, artist.DataHash, artist.CreatedAt,
		)
		artists[i] = artist
	}

	sql := sqlgo.NewSQLGo().SetSQLSchema("public").SetSQLGoInsert(sqlInsert)
	sqlStr := sql.BuildSQL()
	sqlParams := sql.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "artistDAO.Insert", zap.String("SQL", sqlStr), zap.Int("Count", len(artists)))

	_, err := d.dbTrx.GetSqlTx().ExecContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		d.log.Error(ctx, "artistDAO.Insert", zap.Error(err))
		return err
	}
	return nil
}

func (d artistDAO) Update(ctx context.Context, artists entity.Artists) error {
	if len(artists) < 1 {
		return fmt.Errorf("empty artist data")
	}

	for i, artist := range artists {
		now := time.Now()
		artist.UpdatedAt = &now

		socialMediaJSON, err := json.Marshal(artist.ArtistSocialMedia)
		if err != nil {
			socialMediaJSON = []byte("[]")
		}

		sql := sqlgo.NewSQLGo().
			SetSQLSchema("public").
			SetSQLUpdate("artists").
			SetSQLUpdateValue("image", artist.Image).
			SetSQLUpdateValue("name", artist.Name).
			SetSQLUpdateValue("genre", artist.Genre).
			SetSQLUpdateValue("artist_social_media", socialMediaJSON).
			SetSQLUpdateValue("data_hash", artist.DataHash).
			SetSQLUpdateValue("updated_at", artist.UpdatedAt).
			SetSQLWhere("AND", "id", "=", artist.ID)

		sqlStr := sql.BuildSQL()
		sqlParams := sql.GetSQLGoParameter().GetSQLParameter()

		d.log.Debug(ctx, "artistDAO.Update", zap.String("ID", string(artist.ID)))

		_, err = d.dbTrx.GetSqlTx().ExecContext(ctx, sqlStr, sqlParams...)
		if err != nil {
			d.log.Error(ctx, "artistDAO.Update", zap.Error(err))
			return err
		}
		artists[i] = artist
	}
	return nil
}

func (d artistDAO) Delete(ctx context.Context, id pubEntity.UUID) error {
	sql := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLDelete("artists").
		SetSQLWhere("AND", "id", "=", id)

	sqlStr := sql.BuildSQL()
	sqlParams := sql.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "artistDAO.Delete", zap.String("ID", string(id)))

	_, err := d.dbTrx.GetSqlTx().ExecContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		d.log.Error(ctx, "artistDAO.Delete", zap.Error(err))
		return err
	}
	return nil
}

func (d artistDAO) SoftDelete(ctx context.Context, id pubEntity.UUID) error {
	now := time.Now()
	sql := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLUpdate("artists").
		SetSQLUpdateValue("deleted", true).
		SetSQLUpdateValue("updated_at", now).
		SetSQLWhere("AND", "id", "=", id)

	sqlStr := sql.BuildSQL()
	sqlParams := sql.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "artistDAO.SoftDelete", zap.String("ID", string(id)))

	_, err := d.dbTrx.GetSqlTx().ExecContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		d.log.Error(ctx, "artistDAO.SoftDelete", zap.Error(err))
		return err
	}
	return nil
}
