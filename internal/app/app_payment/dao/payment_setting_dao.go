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

type PaymentSettingDAO interface {
	Search(ctx context.Context, query appPayment.PaymentSettingQuery) (appPayment.PaymentSettings, error)
	GetByKey(ctx context.Context, key string) (*appPayment.PaymentSetting, error)
	Update(ctx context.Context, setting *appPayment.PaymentSetting) error
	UpdateSettingValue(ctx context.Context, key string, value bool) error
	UpdateDisplayOrder(ctx context.Context, key string, order int) error
	Upsert(ctx context.Context, key string, value bool, displayOrder int) error
}

type paymentSettingDAO struct {
	log   util.LogUtil
	dbTrx baseDao.DBTransaction
}

func MakePaymentSettingDAO(log util.LogUtil, dbTrx baseDao.DBTransaction) PaymentSettingDAO {
	return &paymentSettingDAO{
		log:   log,
		dbTrx: dbTrx,
	}
}

func (d *paymentSettingDAO) Search(ctx context.Context, query appPayment.PaymentSettingQuery) (appPayment.PaymentSettings, error) {
	sqlSelect := sqlgo.NewSQLGoSelect().
		SetSQLSelect("id", "id").
		SetSQLSelect("setting_key", "setting_key").
		SetSQLSelect("setting_value", "setting_value").
		SetSQLSelect("display_order", "display_order").
		SetSQLSelect("deleted", "deleted").
		SetSQLSelect("data_hash", "data_hash").
		SetSQLSelect("created_at", "created_at").
		SetSQLSelect("updated_at", "updated_at")

	sqlFrom := sqlgo.NewSQLGoFrom().
		SetSQLFrom("payment_settings", "ps")

	sqlWhere := sqlgo.NewSQLGoWhere()
	sqlWhere.SetSQLWhere("AND", "ps.deleted", "=", false)

	if len(query.SettingKeys) > 0 {
		sqlWhere.SetSQLWhere("AND", "ps.setting_key", "IN", query.SettingKeys)
	}

	sqlOrder := sqlgo.NewSQLGoOrder()
	sqlOrder.SetSQLOrder("ps.display_order", "ASC")

	sqlStmt := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLGoSelect(sqlSelect).
		SetSQLGoFrom(sqlFrom).
		SetSQLGoWhere(sqlWhere).
		SetSQLGoOrder(sqlOrder)

	sqlStr := sqlStmt.BuildSQL()
	sqlParams := sqlStmt.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "paymentSettingDAO.Search",
		zap.String("SQL", sqlStr),
		zap.Any("Params", sqlParams),
	)

	rows, err := d.dbTrx.GetSqlDB().QueryContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		d.log.Error(ctx, "paymentSettingDAO.Search", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var settings appPayment.PaymentSettings
	for rows.Next() {
		var setting appPayment.PaymentSetting
		if err := rows.Scan(
			&setting.ID, &setting.SettingKey, &setting.SettingValue,
			&setting.DisplayOrder,
			&setting.Deleted, &setting.DataHash,
			&setting.CreatedAt, &setting.UpdatedAt,
		); err != nil {
			d.log.Error(ctx, "paymentSettingDAO.Search.Scan", zap.Error(err))
			return nil, err
		}
		settings = append(settings, setting)
	}

	return settings, nil
}

func (d *paymentSettingDAO) GetByKey(ctx context.Context, key string) (*appPayment.PaymentSetting, error) {
	sqlSelect := sqlgo.NewSQLGoSelect().
		SetSQLSelect("id", "id").
		SetSQLSelect("setting_key", "setting_key").
		SetSQLSelect("setting_value", "setting_value").
		SetSQLSelect("display_order", "display_order").
		SetSQLSelect("deleted", "deleted").
		SetSQLSelect("data_hash", "data_hash").
		SetSQLSelect("created_at", "created_at").
		SetSQLSelect("updated_at", "updated_at")

	sqlFrom := sqlgo.NewSQLGoFrom().
		SetSQLFrom("payment_settings", "ps")

	sqlWhere := sqlgo.NewSQLGoWhere()
	sqlWhere.SetSQLWhere("AND", "ps.deleted", "=", false)
	sqlWhere.SetSQLWhere("AND", "ps.setting_key", "=", key)

	sqlStmt := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLGoSelect(sqlSelect).
		SetSQLGoFrom(sqlFrom).
		SetSQLGoWhere(sqlWhere)

	sqlStr := sqlStmt.BuildSQL()
	sqlParams := sqlStmt.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "paymentSettingDAO.GetByKey",
		zap.String("SQL", sqlStr),
		zap.Any("Params", sqlParams),
	)

	row := d.dbTrx.GetSqlDB().QueryRowContext(ctx, sqlStr, sqlParams...)

	var setting appPayment.PaymentSetting
	if err := row.Scan(
		&setting.ID, &setting.SettingKey, &setting.SettingValue,
		&setting.DisplayOrder,
		&setting.Deleted, &setting.DataHash,
		&setting.CreatedAt, &setting.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		d.log.Error(ctx, "paymentSettingDAO.GetByKey.Scan", zap.Error(err))
		return nil, err
	}

	return &setting, nil
}

func (d *paymentSettingDAO) Update(ctx context.Context, setting *appPayment.PaymentSetting) error {
	now := time.Now()
	setting.UpdatedAt = &now

	sqlStmt := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLUpdate("payment_settings").
		SetSQLUpdateValue("setting_value", setting.SettingValue).
		SetSQLUpdateValue("display_order", setting.DisplayOrder).
		SetSQLUpdateValue("updated_at", setting.UpdatedAt).
		SetSQLWhere("AND", "id", "=", setting.ID)

	sqlStr := sqlStmt.BuildSQL()
	sqlParams := sqlStmt.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "paymentSettingDAO.Update",
		zap.String("SQL", sqlStr),
		zap.Any("Params", sqlParams),
	)

	_, err := d.dbTrx.GetSqlTx().ExecContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		d.log.Error(ctx, "paymentSettingDAO.Update", zap.Error(err))
		return err
	}

	return nil
}

func (d *paymentSettingDAO) UpdateSettingValue(ctx context.Context, key string, value bool) error {
	now := time.Now()

	sqlStmt := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLUpdate("payment_settings").
		SetSQLUpdateValue("setting_value", value).
		SetSQLUpdateValue("updated_at", now).
		SetSQLWhere("AND", "setting_key", "=", key)

	sqlStr := sqlStmt.BuildSQL()
	sqlParams := sqlStmt.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "paymentSettingDAO.UpdateSettingValue",
		zap.String("SQL", sqlStr),
		zap.Any("Params", sqlParams),
	)

	_, err := d.dbTrx.GetSqlTx().ExecContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		d.log.Error(ctx, "paymentSettingDAO.UpdateSettingValue", zap.Error(err))
		return err
	}

	return nil
}

func (d *paymentSettingDAO) UpdateDisplayOrder(ctx context.Context, key string, order int) error {
	now := time.Now()

	sqlStmt := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLUpdate("payment_settings").
		SetSQLUpdateValue("display_order", order).
		SetSQLUpdateValue("updated_at", now).
		SetSQLWhere("AND", "setting_key", "=", key)

	sqlStr := sqlStmt.BuildSQL()
	sqlParams := sqlStmt.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "paymentSettingDAO.UpdateDisplayOrder",
		zap.String("SQL", sqlStr),
		zap.Any("Params", sqlParams),
	)

	_, err := d.dbTrx.GetSqlTx().ExecContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		d.log.Error(ctx, "paymentSettingDAO.UpdateDisplayOrder", zap.Error(err))
		return err
	}

	return nil
}

func (d *paymentSettingDAO) Upsert(ctx context.Context, key string, value bool, displayOrder int) error {
	now := time.Now()

	sqlStr := `
		INSERT INTO payment_settings (id, setting_key, setting_value, display_order, deleted, data_hash, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, false, '-', $4, $4)
		ON CONFLICT (setting_key) 
		DO UPDATE SET 
			setting_value = $2,
			display_order = $3,
			updated_at = $4
	`

	d.log.Debug(ctx, "paymentSettingDAO.Upsert",
		zap.String("SQL", sqlStr),
		zap.Any("Params", []interface{}{key, value, displayOrder, now}),
	)

	_, err := d.dbTrx.GetSqlTx().ExecContext(ctx, sqlStr, key, value, displayOrder, now)
	if err != nil {
		d.log.Error(ctx, "paymentSettingDAO.Upsert", zap.Error(err))
		return err
	}

	return nil
}
