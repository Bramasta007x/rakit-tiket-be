package dao

import (
	"context"
	"time"

	baseDao "rakit-tiket-be/internal/pkg/dao"
	entity "rakit-tiket-be/pkg/entity/app_gate_log"
	"rakit-tiket-be/pkg/util"

	"gitlab.com/threetopia/sqlgo/v2"
	"go.uber.org/zap"
)

type GateLogDAO interface {
	Search(ctx context.Context, query entity.GateLogQuery) (entity.GateLogs, error)
	Insert(ctx context.Context, logEntry entity.GateLog) error
}

type gateLogDAO struct {
	log   util.LogUtil
	dbTrx baseDao.DBTransaction
}

func MakeGateLogDAO(log util.LogUtil, dbTrx baseDao.DBTransaction) GateLogDAO {
	return gateLogDAO{
		log:   log,
		dbTrx: dbTrx,
	}
}

func (d gateLogDAO) Search(ctx context.Context, query entity.GateLogQuery) (entity.GateLogs, error) {

	sqlSelect := sqlgo.NewSQLGoSelect().
		SetSQLSelect("gl.id", "id").
		SetSQLSelect("gl.event_id", "event_id").
		SetSQLSelect("gl.physical_ticket_id", "physical_ticket_id").
		SetSQLSelect("gl.scanned_by", "scanned_by").
		SetSQLSelect("gl.action", "action").
		SetSQLSelect("gl.success", "success").
		SetSQLSelect("gl.message", "message").
		SetSQLSelect("gl.gate_name", "gate_name").
		SetSQLSelect("gl.ticket_type", "ticket_type").
		SetSQLSelect("gl.scan_sequence", "scan_sequence").
		SetSQLSelect("gl.created_at", "created_at")

	sqlFrom := sqlgo.NewSQLGoFrom().
		SetSQLFrom("gate_logs", "gl")

	sqlWhere := sqlgo.NewSQLGoWhere()

	if len(query.IDs) > 0 {
		sqlWhere.SetSQLWhere("AND", "gl.id", "IN", query.IDs)
	}
	if len(query.EventIDs) > 0 {
		sqlWhere.SetSQLWhere("AND", "gl.event_id", "IN", query.EventIDs)
	}
	if len(query.PhysicalTicketIDs) > 0 {
		sqlWhere.SetSQLWhere("AND", "gl.physical_ticket_id", "IN", query.PhysicalTicketIDs)
	}
	if len(query.Actions) > 0 {
		sqlWhere.SetSQLWhere("AND", "gl.action", "IN", query.Actions)
	}
	if len(query.GateNames) > 0 {
		sqlWhere.SetSQLWhere("AND", "gl.gate_name", "IN", query.GateNames)
	}

	sql := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLGoSelect(sqlSelect).
		SetSQLGoFrom(sqlFrom).
		SetSQLGoWhere(sqlWhere)

	sqlStr := sql.BuildSQL()
	sqlParams := sql.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "gateLogDAO.Search",
		zap.String("SQL", sqlStr),
		zap.Any("Params", sqlParams),
	)

	rows, err := d.dbTrx.GetSqlDB().QueryContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		d.log.Error(ctx, "gateLogDAO.Search",
			zap.String("SQL", sqlStr),
			zap.Any("Params", sqlParams),
			zap.Error(err),
		)
		return nil, err
	}
	defer rows.Close()

	var logs entity.GateLogs
	for rows.Next() {
		var gl entity.GateLog

		if err := rows.Scan(
			&gl.ID,
			&gl.EventID,
			&gl.PhysicalTicketID,
			&gl.ScannedBy,
			&gl.Action,
			&gl.Success,
			&gl.Message,
			&gl.GateName,
			&gl.TicketType,
			&gl.ScanSequence,
			&gl.CreatedAt,
		); err != nil {
			d.log.Error(ctx, "gateLogDAO.Search.Scan", zap.Error(err))
			return nil, err
		}

		logs = append(logs, gl)
	}

	return logs, nil
}

func (d gateLogDAO) Insert(ctx context.Context, logEntry entity.GateLog) error {

	logEntry.CreatedAt = time.Now()

	sql := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLGoInsert(sqlgo.NewSQLGoInsert()).
		SetSQLInsert("gate_logs").
		SetSQLInsertColumn(
			"id",
			"event_id",
			"physical_ticket_id",
			"scanned_by",
			"action",
			"success",
			"message",
			"gate_name",
			"ticket_type",
			"scan_sequence",
			"created_at",
		).
		SetSQLInsertValue(
			logEntry.ID,
			logEntry.EventID,
			logEntry.PhysicalTicketID,
			logEntry.ScannedBy,
			logEntry.Action,
			logEntry.Success,
			logEntry.Message,
			logEntry.GateName,
			logEntry.TicketType,
			logEntry.ScanSequence,
			logEntry.CreatedAt,
		)

	sqlStr := sql.BuildSQL()
	sqlParams := sql.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "gateLogDAO.Insert",
		zap.String("SQL", sqlStr),
		zap.Any("Params", sqlParams),
	)

	_, err := d.dbTrx.GetSqlTx().ExecContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		d.log.Error(ctx, "gateLogDAO.Insert",
			zap.String("SQL", sqlStr),
			zap.Any("Params", sqlParams),
			zap.Error(err),
		)
		return err
	}

	return nil
}
