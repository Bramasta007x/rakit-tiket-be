package dao

import (
	"context"
	"encoding/json"
	"time"

	baseDao "rakit-tiket-be/internal/pkg/dao"
	pubEntity "rakit-tiket-be/pkg/entity"
	entity "rakit-tiket-be/pkg/entity/app_gate_config"
	"rakit-tiket-be/pkg/util"

	"gitlab.com/threetopia/sqlgo/v2"
	"go.uber.org/zap"
)

type GateConfigDAO interface {
	Search(ctx context.Context, query entity.GateConfigQuery) (entity.GateConfigs, error)
	GetByEventID(ctx context.Context, eventID string) (*entity.GateConfig, error)
	Insert(ctx context.Context, config entity.GateConfig) error
	Update(ctx context.Context, config entity.GateConfig) error
	Delete(ctx context.Context, id pubEntity.UUID) error
}

type gateConfigDAO struct {
	log   util.LogUtil
	dbTrx baseDao.DBTransaction
}

func MakeGateConfigDAO(log util.LogUtil, dbTrx baseDao.DBTransaction) GateConfigDAO {
	return gateConfigDAO{
		log:   log,
		dbTrx: dbTrx,
	}
}

func (d gateConfigDAO) Search(ctx context.Context, query entity.GateConfigQuery) (entity.GateConfigs, error) {

	sqlSelect := sqlgo.NewSQLGoSelect().
		SetSQLSelect("gc.id", "id").
		SetSQLSelect("gc.event_id", "event_id").
		SetSQLSelect("gc.mode", "mode").
		SetSQLSelect("gc.max_scan_per_ticket", "max_scan_per_ticket").
		SetSQLSelect("gc.max_scan_by_type", "max_scan_by_type").
		SetSQLSelect("gc.is_active", "is_active").
		SetSQLSelect("gc.created_at", "created_at").
		SetSQLSelect("gc.updated_at", "updated_at")

	sqlFrom := sqlgo.NewSQLGoFrom().
		SetSQLFrom("gate_configs", "gc")

	sqlWhere := sqlgo.NewSQLGoWhere().
		SetSQLWhere("AND", "gc.deleted", "=", false)

	if len(query.IDs) > 0 {
		sqlWhere.SetSQLWhere("AND", "gc.id", "IN", query.IDs)
	}
	if len(query.EventIDs) > 0 {
		sqlWhere.SetSQLWhere("AND", "gc.event_id", "IN", query.EventIDs)
	}
	if query.IsActive != nil {
		sqlWhere.SetSQLWhere("AND", "gc.is_active", "=", *query.IsActive)
	}

	sql := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLGoSelect(sqlSelect).
		SetSQLGoFrom(sqlFrom).
		SetSQLGoWhere(sqlWhere)

	sqlStr := sql.BuildSQL()
	sqlParams := sql.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "gateConfigDAO.Search",
		zap.String("SQL", sqlStr),
		zap.Any("Params", sqlParams),
	)

	rows, err := d.dbTrx.GetSqlDB().QueryContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		d.log.Error(ctx, "gateConfigDAO.Search",
			zap.String("SQL", sqlStr),
			zap.Any("Params", sqlParams),
			zap.Error(err),
		)
		return nil, err
	}
	defer rows.Close()

	var configs entity.GateConfigs
	for rows.Next() {
		var cfg entity.GateConfig
		var maxScanByTypeJSON []byte

		if err := rows.Scan(
			&cfg.ID,
			&cfg.EventID,
			&cfg.Mode,
			&cfg.MaxScanPerTicket,
			&maxScanByTypeJSON,
			&cfg.IsActive,
			&cfg.CreatedAt,
			&cfg.UpdatedAt,
		); err != nil {
			d.log.Error(ctx, "gateConfigDAO.Search.Scan", zap.Error(err))
			return nil, err
		}

		if len(maxScanByTypeJSON) > 0 {
			_ = json.Unmarshal(maxScanByTypeJSON, &cfg.MaxScanByType)
		}

		configs = append(configs, cfg)
	}

	return configs, nil
}

func (d gateConfigDAO) GetByEventID(ctx context.Context, eventID string) (*entity.GateConfig, error) {

	sqlSelect := sqlgo.NewSQLGoSelect().
		SetSQLSelect("gc.id", "id").
		SetSQLSelect("gc.event_id", "event_id").
		SetSQLSelect("gc.mode", "mode").
		SetSQLSelect("gc.max_scan_per_ticket", "max_scan_per_ticket").
		SetSQLSelect("gc.max_scan_by_type", "max_scan_by_type").
		SetSQLSelect("gc.is_active", "is_active").
		SetSQLSelect("gc.created_at", "created_at").
		SetSQLSelect("gc.updated_at", "updated_at")

	sqlFrom := sqlgo.NewSQLGoFrom().
		SetSQLFrom("gate_configs", "gc")

	sqlWhere := sqlgo.NewSQLGoWhere().
		SetSQLWhere("AND", "gc.deleted", "=", false).
		SetSQLWhere("AND", "gc.event_id", "=", eventID)

	sql := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLGoSelect(sqlSelect).
		SetSQLGoFrom(sqlFrom).
		SetSQLGoWhere(sqlWhere)

	sqlStr := sql.BuildSQL()
	sqlParams := sql.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "gateConfigDAO.GetByEventID",
		zap.String("SQL", sqlStr),
		zap.Any("Params", sqlParams),
	)

	var cfg entity.GateConfig
	var maxScanByTypeJSON []byte

	err := d.dbTrx.GetSqlDB().QueryRowContext(ctx, sqlStr, sqlParams...).Scan(
		&cfg.ID,
		&cfg.EventID,
		&cfg.Mode,
		&cfg.MaxScanPerTicket,
		&maxScanByTypeJSON,
		&cfg.IsActive,
		&cfg.CreatedAt,
		&cfg.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if len(maxScanByTypeJSON) > 0 {
		_ = json.Unmarshal(maxScanByTypeJSON, &cfg.MaxScanByType)
	}

	return &cfg, nil
}

func (d gateConfigDAO) Insert(ctx context.Context, config entity.GateConfig) error {

	now := time.Now()
	config.CreatedAt = now
	config.UpdatedAt = &now

	var maxScanByTypeJSON []byte
	if config.MaxScanByType != nil {
		maxScanByTypeJSON, _ = json.Marshal(config.MaxScanByType)
	}

	sql := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLGoInsert(sqlgo.NewSQLGoInsert()).
		SetSQLInsert("gate_configs").
		SetSQLInsertColumn(
			"id",
			"event_id",
			"mode",
			"max_scan_per_ticket",
			"max_scan_by_type",
			"is_active",
			"data_hash",
			"created_at",
		).
		SetSQLInsertValue(
			config.ID,
			config.EventID,
			config.Mode,
			config.MaxScanPerTicket,
			maxScanByTypeJSON,
			config.IsActive,
			config.DataHash,
			config.CreatedAt,
		)

	sqlStr := sql.BuildSQL()
	sqlParams := sql.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "gateConfigDAO.Insert",
		zap.String("SQL", sqlStr),
		zap.Any("Params", sqlParams),
	)

	_, err := d.dbTrx.GetSqlTx().ExecContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		d.log.Error(ctx, "gateConfigDAO.Insert",
			zap.String("SQL", sqlStr),
			zap.Any("Params", sqlParams),
			zap.Error(err),
		)
		return err
	}

	return nil
}

func (d gateConfigDAO) Update(ctx context.Context, config entity.GateConfig) error {

	now := time.Now()
	config.UpdatedAt = &now

	var maxScanByTypeJSON []byte
	if config.MaxScanByType != nil {
		maxScanByTypeJSON, _ = json.Marshal(config.MaxScanByType)
	}

	sql := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLUpdate("gate_configs").
		SetSQLUpdateValue("mode", config.Mode).
		SetSQLUpdateValue("max_scan_per_ticket", config.MaxScanPerTicket).
		SetSQLUpdateValue("max_scan_by_type", maxScanByTypeJSON).
		SetSQLUpdateValue("is_active", config.IsActive).
		SetSQLUpdateValue("updated_at", config.UpdatedAt).
		SetSQLWhere("AND", "id", "=", config.ID)

	sqlStr := sql.BuildSQL()
	sqlParams := sql.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "gateConfigDAO.Update",
		zap.String("SQL", sqlStr),
		zap.Any("Params", sqlParams),
	)

	_, err := d.dbTrx.GetSqlTx().ExecContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		d.log.Error(ctx, "gateConfigDAO.Update",
			zap.String("SQL", sqlStr),
			zap.Any("Params", sqlParams),
			zap.Error(err),
		)
		return err
	}

	return nil
}

func (d gateConfigDAO) Delete(ctx context.Context, id pubEntity.UUID) error {

	sql := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLDelete("gate_configs").
		SetSQLWhere("AND", "id", "=", id)

	sqlStr := sql.BuildSQL()
	sqlParams := sql.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "gateConfigDAO.Delete",
		zap.String("SQL", sqlStr),
		zap.Any("Params", sqlParams),
	)

	_, err := d.dbTrx.GetSqlTx().ExecContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		d.log.Error(ctx, "gateConfigDAO.Delete",
			zap.String("SQL", sqlStr),
			zap.Any("Params", sqlParams),
			zap.Error(err),
		)
		return err
	}

	return nil
}
