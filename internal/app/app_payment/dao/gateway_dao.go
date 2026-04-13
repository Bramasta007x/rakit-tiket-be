package dao

import (
	"context"
	"database/sql"
	"time"

	baseDao "rakit-tiket-be/internal/pkg/dao"
	appPayment "rakit-tiket-be/pkg/entity/app_payment"
	"rakit-tiket-be/pkg/util"

	"gitlab.com/threetopia/sqlgo/v2"
	"go.uber.org/zap"
)

type GatewayDAO interface {
	Search(ctx context.Context, query appPayment.GatewayQuery) (appPayment.PaymentGateways, error)
	GetByID(ctx context.Context, id string) (*appPayment.PaymentGateway, error)
	Update(ctx context.Context, gateway *appPayment.PaymentGateway) error
	UpdateDisplayOrder(ctx context.Context, code string, order int) error
	SetActiveGateway(ctx context.Context, code string) error
	DeactivateAll(ctx context.Context) error
}

type gatewayDAO struct {
	log   util.LogUtil
	dbTrx baseDao.DBTransaction
}

func MakeGatewayDAO(log util.LogUtil, dbTrx baseDao.DBTransaction) GatewayDAO {
	return &gatewayDAO{
		log:   log,
		dbTrx: dbTrx,
	}
}

func (d *gatewayDAO) Search(ctx context.Context, query appPayment.GatewayQuery) (appPayment.PaymentGateways, error) {
	sqlSelect := sqlgo.NewSQLGoSelect().
		SetSQLSelect("id", "id").
		SetSQLSelect("code", "code").
		SetSQLSelect("name", "name").
		SetSQLSelect("is_enabled", "is_enabled").
		SetSQLSelect("is_active", "is_active").
		SetSQLSelect("display_order", "display_order").
		SetSQLSelect("deleted", "deleted").
		SetSQLSelect("data_hash", "data_hash").
		SetSQLSelect("created_at", "created_at").
		SetSQLSelect("updated_at", "updated_at")

	sqlFrom := sqlgo.NewSQLGoFrom().
		SetSQLFrom("payment_gateways", "pg")

	sqlWhere := sqlgo.NewSQLGoWhere()
	sqlWhere.SetSQLWhere("AND", "pg.deleted", "=", false)

	if len(query.IDs) > 0 {
		sqlWhere.SetSQLWhere("AND", "pg.id", "IN", query.IDs)
	}

	if len(query.Codes) > 0 {
		sqlWhere.SetSQLWhere("AND", "pg.code", "IN", query.Codes)
	}

	if len(query.Names) > 0 {
		sqlWhere.SetSQLWhere("AND", "pg.name", "IN", query.Names)
	}

	if query.IsEnabled != nil {
		sqlWhere.SetSQLWhere("AND", "pg.is_enabled", "=", *query.IsEnabled)
	}

	if query.IsActive != nil {
		sqlWhere.SetSQLWhere("AND", "pg.is_active", "=", *query.IsActive)
	}

	sqlOrder := sqlgo.NewSQLGoOrder()
	sqlOrder.SetSQLOrder("pg.display_order", "ASC")

	sqlStmt := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLGoSelect(sqlSelect).
		SetSQLGoFrom(sqlFrom).
		SetSQLGoWhere(sqlWhere).
		SetSQLGoOrder(sqlOrder)

	sqlStr := sqlStmt.BuildSQL()
	sqlParams := sqlStmt.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "gatewayDAO.Search",
		zap.String("SQL", sqlStr),
		zap.Any("Params", sqlParams),
	)

	rows, err := d.dbTrx.GetSqlDB().QueryContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		d.log.Error(ctx, "gatewayDAO.Search", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var gateways appPayment.PaymentGateways
	for rows.Next() {
		var gateway appPayment.PaymentGateway
		if err := rows.Scan(
			&gateway.ID, &gateway.Code, &gateway.Name,
			&gateway.IsEnabled, &gateway.IsActive, &gateway.DisplayOrder,
			&gateway.Deleted, &gateway.DataHash,
			&gateway.CreatedAt, &gateway.UpdatedAt,
		); err != nil {
			d.log.Error(ctx, "gatewayDAO.Search.Scan", zap.Error(err))
			return nil, err
		}
		gateways = append(gateways, gateway)
	}

	return gateways, nil
}

func (d *gatewayDAO) GetByID(ctx context.Context, id string) (*appPayment.PaymentGateway, error) {
	sqlSelect := sqlgo.NewSQLGoSelect().
		SetSQLSelect("id", "id").
		SetSQLSelect("code", "code").
		SetSQLSelect("name", "name").
		SetSQLSelect("is_enabled", "is_enabled").
		SetSQLSelect("is_active", "is_active").
		SetSQLSelect("display_order", "display_order").
		SetSQLSelect("deleted", "deleted").
		SetSQLSelect("data_hash", "data_hash").
		SetSQLSelect("created_at", "created_at").
		SetSQLSelect("updated_at", "updated_at")

	sqlFrom := sqlgo.NewSQLGoFrom().
		SetSQLFrom("payment_gateways", "pg")

	sqlWhere := sqlgo.NewSQLGoWhere()
	sqlWhere.SetSQLWhere("AND", "pg.deleted", "=", false)
	sqlWhere.SetSQLWhere("AND", "pg.id", "=", id)

	sqlStmt := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLGoSelect(sqlSelect).
		SetSQLGoFrom(sqlFrom).
		SetSQLGoWhere(sqlWhere)

	sqlStr := sqlStmt.BuildSQL()
	sqlParams := sqlStmt.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "gatewayDAO.GetByID",
		zap.String("SQL", sqlStr),
		zap.Any("Params", sqlParams),
	)

	row := d.dbTrx.GetSqlDB().QueryRowContext(ctx, sqlStr, sqlParams...)

	var gateway appPayment.PaymentGateway
	if err := row.Scan(
		&gateway.ID, &gateway.Code, &gateway.Name,
		&gateway.IsEnabled, &gateway.IsActive, &gateway.DisplayOrder,
		&gateway.Deleted, &gateway.DataHash,
		&gateway.CreatedAt, &gateway.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		d.log.Error(ctx, "gatewayDAO.GetByID.Scan", zap.Error(err))
		return nil, err
	}

	return &gateway, nil
}

func (d *gatewayDAO) Update(ctx context.Context, gateway *appPayment.PaymentGateway) error {
	now := time.Now()
	gateway.UpdatedAt = &now

	sqlStmt := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLUpdate("payment_gateways").
		SetSQLUpdateValue("is_enabled", gateway.IsEnabled).
		SetSQLUpdateValue("is_active", gateway.IsActive).
		SetSQLUpdateValue("display_order", gateway.DisplayOrder).
		SetSQLUpdateValue("updated_at", gateway.UpdatedAt).
		SetSQLWhere("AND", "id", "=", gateway.ID)

	sqlStr := sqlStmt.BuildSQL()
	sqlParams := sqlStmt.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "gatewayDAO.Update",
		zap.String("SQL", sqlStr),
		zap.Any("Params", sqlParams),
	)

	_, err := d.dbTrx.GetSqlTx().ExecContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		d.log.Error(ctx, "gatewayDAO.Update", zap.Error(err))
		return err
	}

	return nil
}

func (d *gatewayDAO) UpdateDisplayOrder(ctx context.Context, code string, order int) error {
	now := time.Now()

	sqlStmt := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLUpdate("payment_gateways").
		SetSQLUpdateValue("display_order", order).
		SetSQLUpdateValue("updated_at", now).
		SetSQLWhere("AND", "code", "=", code)

	sqlStr := sqlStmt.BuildSQL()
	sqlParams := sqlStmt.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "gatewayDAO.UpdateDisplayOrder",
		zap.String("SQL", sqlStr),
		zap.Any("Params", sqlParams),
	)

	_, err := d.dbTrx.GetSqlTx().ExecContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		d.log.Error(ctx, "gatewayDAO.UpdateDisplayOrder", zap.Error(err))
		return err
	}

	return nil
}

func (d *gatewayDAO) SetActiveGateway(ctx context.Context, code string) error {
	now := time.Now()

	sqlStmt := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLUpdate("payment_gateways").
		SetSQLUpdateValue("is_active", true).
		SetSQLUpdateValue("updated_at", now).
		SetSQLWhere("AND", "code", "=", code)

	sqlStr := sqlStmt.BuildSQL()
	sqlParams := sqlStmt.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "gatewayDAO.SetActiveGateway",
		zap.String("SQL", sqlStr),
		zap.Any("Params", sqlParams),
	)

	_, err := d.dbTrx.GetSqlTx().ExecContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		d.log.Error(ctx, "gatewayDAO.SetActiveGateway", zap.Error(err))
		return err
	}

	return nil
}

func (d *gatewayDAO) DeactivateAll(ctx context.Context) error {
	now := time.Now()

	sqlStmt := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLUpdate("payment_gateways").
		SetSQLUpdateValue("is_active", false).
		SetSQLUpdateValue("updated_at", now)

	sqlStr := sqlStmt.BuildSQL()
	sqlParams := sqlStmt.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "gatewayDAO.DeactivateAll",
		zap.String("SQL", sqlStr),
		zap.Any("Params", sqlParams),
	)

	_, err := d.dbTrx.GetSqlTx().ExecContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		d.log.Error(ctx, "gatewayDAO.DeactivateAll", zap.Error(err))
		return err
	}

	return nil
}
