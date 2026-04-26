package dao

import (
	"context"
	"time"

	baseDao "rakit-tiket-be/internal/pkg/dao"
	pubEntity "rakit-tiket-be/pkg/entity"
	entity "rakit-tiket-be/pkg/entity/app_physical_ticket"
	"rakit-tiket-be/pkg/util"

	"gitlab.com/threetopia/sqlgo/v2"
	"go.uber.org/zap"
)

type PhysicalTicketDAO interface {
	Search(ctx context.Context, query entity.PhysicalTicketQuery) (entity.PhysicalTickets, error)
	SearchByQRCode(ctx context.Context, qrCode string) (*entity.PhysicalTicket, error)
	Insert(ctx context.Context, tickets entity.PhysicalTickets) error
	Update(ctx context.Context, tickets entity.PhysicalTickets) error
	SoftDelete(ctx context.Context, id pubEntity.UUID) error
}

type physicalTicketDAO struct {
	log   util.LogUtil
	dbTrx baseDao.DBTransaction
}

func MakePhysicalTicketDAO(log util.LogUtil, dbTrx baseDao.DBTransaction) PhysicalTicketDAO {
	return physicalTicketDAO{
		log:   log,
		dbTrx: dbTrx,
	}
}

func (d physicalTicketDAO) Search(ctx context.Context, query entity.PhysicalTicketQuery) (entity.PhysicalTickets, error) {

	sqlSelect := sqlgo.NewSQLGoSelect().
		SetSQLSelect("pt.id", "id").
		SetSQLSelect("pt.event_id", "event_id").
		SetSQLSelect("pt.ticket_type", "ticket_type").
		SetSQLSelect("pt.ticket_id", "ticket_id").
		SetSQLSelect("pt.qr_code", "qr_code").
		SetSQLSelect("pt.qr_code_hash", "qr_code_hash").
		SetSQLSelect("pt.status", "status").
		SetSQLSelect("pt.scan_count", "scan_count").
		SetSQLSelect("pt.checked_in_at", "checked_in_at").
		SetSQLSelect("pt.checked_out_at", "checked_out_at").
		SetSQLSelect("pt.created_at", "created_at").
		SetSQLSelect("pt.updated_at", "updated_at")

	sqlFrom := sqlgo.NewSQLGoFrom().
		SetSQLFrom("physical_tickets", "pt")

	sqlWhere := sqlgo.NewSQLGoWhere().
		SetSQLWhere("AND", "pt.deleted", "=", false)

	if len(query.IDs) > 0 {
		sqlWhere.SetSQLWhere("AND", "pt.id", "IN", query.IDs)
	}
	if len(query.EventIDs) > 0 {
		sqlWhere.SetSQLWhere("AND", "pt.event_id", "IN", query.EventIDs)
	}
	if len(query.TicketIDs) > 0 {
		sqlWhere.SetSQLWhere("AND", "pt.ticket_id", "IN", query.TicketIDs)
	}
	if len(query.TicketTypes) > 0 {
		sqlWhere.SetSQLWhere("AND", "pt.ticket_type", "IN", query.TicketTypes)
	}
	if len(query.QRCodes) > 0 {
		sqlWhere.SetSQLWhere("AND", "pt.qr_code", "IN", query.QRCodes)
	}
	if len(query.Statuses) > 0 {
		sqlWhere.SetSQLWhere("AND", "pt.status", "IN", query.Statuses)
	}

	sql := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLGoSelect(sqlSelect).
		SetSQLGoFrom(sqlFrom).
		SetSQLGoWhere(sqlWhere)

	sqlStr := sql.BuildSQL()
	sqlParams := sql.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "physicalTicketDAO.Search",
		zap.String("SQL", sqlStr),
		zap.Any("Params", sqlParams),
	)

	rows, err := d.dbTrx.GetSqlDB().QueryContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		d.log.Error(ctx, "physicalTicketDAO.Search",
			zap.String("SQL", sqlStr),
			zap.Any("Params", sqlParams),
			zap.Error(err),
		)
		return nil, err
	}
	defer rows.Close()

	var tickets entity.PhysicalTickets
	for rows.Next() {
		var tkt entity.PhysicalTicket

		if err := rows.Scan(
			&tkt.ID,
			&tkt.EventID,
			&tkt.TicketType,
			&tkt.TicketID,
			&tkt.QRCode,
			&tkt.QRCodeHash,
			&tkt.Status,
			&tkt.ScanCount,
			&tkt.CheckedInAt,
			&tkt.CheckedOutAt,
			&tkt.CreatedAt,
			&tkt.UpdatedAt,
		); err != nil {
			d.log.Error(ctx, "physicalTicketDAO.Search.Scan", zap.Error(err))
			return nil, err
		}

		tickets = append(tickets, tkt)
	}

	return tickets, nil
}

func (d physicalTicketDAO) SearchByQRCode(ctx context.Context, qrCode string) (*entity.PhysicalTicket, error) {

	sqlSelect := sqlgo.NewSQLGoSelect().
		SetSQLSelect("pt.id", "id").
		SetSQLSelect("pt.event_id", "event_id").
		SetSQLSelect("pt.ticket_type", "ticket_type").
		SetSQLSelect("pt.ticket_id", "ticket_id").
		SetSQLSelect("pt.qr_code", "qr_code").
		SetSQLSelect("pt.qr_code_hash", "qr_code_hash").
		SetSQLSelect("pt.status", "status").
		SetSQLSelect("pt.scan_count", "scan_count").
		SetSQLSelect("pt.checked_in_at", "checked_in_at").
		SetSQLSelect("pt.checked_out_at", "checked_out_at").
		SetSQLSelect("pt.created_at", "created_at").
		SetSQLSelect("pt.updated_at", "updated_at")

	sqlFrom := sqlgo.NewSQLGoFrom().
		SetSQLFrom("physical_tickets", "pt")

	sqlWhere := sqlgo.NewSQLGoWhere().
		SetSQLWhere("AND", "pt.deleted", "=", false).
		SetSQLWhere("AND", "pt.qr_code", "=", qrCode)

	sql := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLGoSelect(sqlSelect).
		SetSQLGoFrom(sqlFrom).
		SetSQLGoWhere(sqlWhere)

	sqlStr := sql.BuildSQL()
	sqlParams := sql.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "physicalTicketDAO.SearchByQRCode",
		zap.String("SQL", sqlStr),
		zap.Any("Params", sqlParams),
	)

	var tkt entity.PhysicalTicket
	err := d.dbTrx.GetSqlDB().QueryRowContext(ctx, sqlStr, sqlParams...).Scan(
		&tkt.ID,
		&tkt.EventID,
		&tkt.TicketType,
		&tkt.TicketID,
		&tkt.QRCode,
		&tkt.QRCodeHash,
		&tkt.Status,
		&tkt.ScanCount,
		&tkt.CheckedInAt,
		&tkt.CheckedOutAt,
		&tkt.CreatedAt,
		&tkt.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &tkt, nil
}

func (d physicalTicketDAO) Insert(ctx context.Context, tickets entity.PhysicalTickets) error {

	if len(tickets) < 1 {
		return nil
	}

	sqlInsert := sqlgo.NewSQLGoInsert().
		SetSQLInsert("physical_tickets").
		SetSQLInsertColumn(
			"id",
			"event_id",
			"ticket_type",
			"ticket_id",
			"qr_code",
			"qr_code_hash",
			"status",
			"scan_count",
			"data_hash",
			"created_at",
		)

	for i, tkt := range tickets {
		tkt.CreatedAt = time.Now()
		tkt.Status = entity.PhysicalTicketStatusActive
		tkt.ScanCount = 0

		sqlInsert.SetSQLInsertValue(
			tkt.ID,
			tkt.EventID,
			tkt.TicketType,
			tkt.TicketID,
			tkt.QRCode,
			tkt.QRCodeHash,
			tkt.Status,
			tkt.ScanCount,
			tkt.DaoEntity.DataHash,
			tkt.CreatedAt,
		)

		tickets[i] = tkt
	}

	sql := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLGoInsert(sqlInsert)

	sqlStr := sql.BuildSQL()
	sqlParams := sql.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "physicalTicketDAO.Insert",
		zap.String("SQL", sqlStr),
		zap.Any("Params", sqlParams),
		zap.Int("Len", len(sqlParams)),
	)

	_, err := d.dbTrx.GetSqlTx().ExecContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		d.log.Error(ctx, "physicalTicketDAO.Insert",
			zap.String("SQL", sqlStr),
			zap.Any("Params", sqlParams),
			zap.Error(err),
		)
		return err
	}

	return nil
}

func (d physicalTicketDAO) Update(ctx context.Context, tickets entity.PhysicalTickets) error {
	for _, tkt := range tickets {
		now := time.Now()
		tkt.UpdatedAt = &now

		sql := sqlgo.NewSQLGo().
			SetSQLSchema("public").
			SetSQLUpdate("physical_tickets").
			SetSQLUpdateValue("event_id", tkt.EventID).
			SetSQLUpdateValue("ticket_type", tkt.TicketType).
			SetSQLUpdateValue("ticket_id", tkt.TicketID).
			SetSQLUpdateValue("qr_code", tkt.QRCode).
			SetSQLUpdateValue("qr_code_hash", tkt.QRCodeHash).
			SetSQLUpdateValue("status", tkt.Status).
			SetSQLUpdateValue("scan_count", tkt.ScanCount).
			SetSQLUpdateValue("checked_in_at", tkt.CheckedInAt).
			SetSQLUpdateValue("checked_out_at", tkt.CheckedOutAt).
			SetSQLUpdateValue("updated_at", tkt.UpdatedAt).
			SetSQLWhere("AND", "id", "=", tkt.ID)

		_, err := d.dbTrx.GetSqlTx().ExecContext(ctx, sql.BuildSQL(), sql.GetSQLGoParameter().GetSQLParameter()...)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d physicalTicketDAO) SoftDelete(ctx context.Context, id pubEntity.UUID) error {

	sql := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLUpdate("physical_tickets").
		SetSQLUpdateValue("deleted", true).
		SetSQLWhere("AND", "id", "=", id)

	sqlStr := sql.BuildSQL()
	sqlParams := sql.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "physicalTicketDAO.SoftDelete",
		zap.String("SQL", sqlStr),
		zap.Any("Params", sqlParams),
	)

	_, err := d.dbTrx.GetSqlTx().ExecContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		d.log.Error(ctx, "physicalTicketDAO.SoftDelete",
			zap.String("SQL", sqlStr),
			zap.Any("Params", sqlParams),
			zap.Error(err),
		)
		return err
	}

	return nil
}
