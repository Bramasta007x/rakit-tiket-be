package dao

import (
	"context"
	"fmt"
	"time"

	baseDao "rakit-tiket-be/internal/pkg/dao"
	pubEntity "rakit-tiket-be/pkg/entity"
	entity "rakit-tiket-be/pkg/entity/app_order"
	"rakit-tiket-be/pkg/util"

	"gitlab.com/threetopia/sqlgo/v2"
	"go.uber.org/zap"
)

type OrderDAO interface {
	Search(ctx context.Context, query entity.OrderQuery) (entity.Orders, error)
	SearchForUpdate(ctx context.Context, query entity.OrderQuery) (entity.Orders, error)
	Insert(ctx context.Context, orders entity.Orders) error
	Update(ctx context.Context, orders entity.Orders) error
	Delete(ctx context.Context, id pubEntity.UUID) error
	SoftDelete(ctx context.Context, id pubEntity.UUID) error
}

type orderDAO struct {
	log   util.LogUtil
	dbTrx baseDao.DBTransaction
}

func MakeOrderDAO(log util.LogUtil, dbTrx baseDao.DBTransaction) OrderDAO {
	return orderDAO{
		log:   log,
		dbTrx: dbTrx,
	}
}

func (d orderDAO) Search(ctx context.Context, query entity.OrderQuery) (entity.Orders, error) {

	sqlSelect := sqlgo.NewSQLGoSelect().
		SetSQLSelect("o.id", "id").
		SetSQLSelect("o.event_id", "event_id").
		SetSQLSelect("o.registrant_id", "registrant_id").
		SetSQLSelect("o.order_number", "order_number").
		SetSQLSelect("o.amount", "amount").
		SetSQLSelect("o.currency", "currency").
		SetSQLSelect("o.payment_type", "payment_type").
		SetSQLSelect("o.payment_gateway", "payment_gateway").
		SetSQLSelect("o.payment_method", "payment_method").
		SetSQLSelect("o.payment_channel", "payment_channel").
		SetSQLSelect("o.payment_status", "payment_status").
		SetSQLSelect("o.payment_token", "payment_token").
		SetSQLSelect("o.payment_url", "payment_url").
		SetSQLSelect("o.payment_transaction_id", "payment_transaction_id").
		SetSQLSelect("o.payment_metadata", "payment_metadata").
		SetSQLSelect("o.payment_proof_url", "payment_proof_url").
		SetSQLSelect("o.payment_proof_filename", "payment_proof_filename").
		SetSQLSelect("o.verified_by", "verified_by").
		SetSQLSelect("o.verified_at", "verified_at").
		SetSQLSelect("o.payment_time", "payment_time").
		SetSQLSelect("o.expires_at", "expires_at").
		SetSQLSelect("o.deleted", "deleted").
		SetSQLSelect("o.data_hash", "data_hash").
		SetSQLSelect("o.created_at", "created_at").
		SetSQLSelect("o.updated_at", "updated_at")

	sqlFrom := sqlgo.NewSQLGoFrom().
		SetSQLFrom("orders", "o")

	sqlWhere := sqlgo.NewSQLGoWhere()
	sqlWhere.SetSQLWhere("AND", "o.deleted", "=", false)

	if len(query.IDs) > 0 {
		sqlWhere.SetSQLWhere("AND", "o.id", "IN", query.IDs)
	}

	if len(query.RegistrantIDs) > 0 {
		sqlWhere.SetSQLWhere("AND", "o.registrant_id", "IN", query.RegistrantIDs)
	}

	if len(query.EventIDs) > 0 {
		sqlWhere.SetSQLWhere("AND", "o.event_id", "IN", query.EventIDs)
	}

	if len(query.OrderNumbers) > 0 {
		sqlWhere.SetSQLWhere("AND", "o.order_number", "IN", query.OrderNumbers)
	}

	if len(query.Statuses) > 0 {
		sqlWhere.SetSQLWhere("AND", "o.payment_status", "IN", query.Statuses)
	}

	if len(query.PaymentGateways) > 0 {
		sqlWhere.SetSQLWhere("AND", "o.payment_gateway", "IN", query.PaymentGateways)
	}

	if query.ExpiredBefore != nil {
		sqlWhere.SetSQLWhere("AND", "o.expires_at", "<", query.ExpiredBefore)
	}

	sql := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLGoSelect(sqlSelect).
		SetSQLGoFrom(sqlFrom).
		SetSQLGoWhere(sqlWhere)

	sqlStr := sql.BuildSQL()
	sqlParams := sql.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "orderDAO.Search",
		zap.String("SQL", sqlStr),
		zap.Any("Params", sqlParams),
	)

	rows, err := d.dbTrx.GetSqlDB().QueryContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		d.log.Error(ctx, "orderDAO.Search",
			zap.String("SQL", sqlStr),
			zap.Any("Params", sqlParams),
			zap.Error(err),
		)
		return nil, err
	}
	defer rows.Close()

	var orders entity.Orders
	for rows.Next() {
		var order entity.Order

		if err := rows.Scan(
			&order.ID,
			&order.EventID,
			&order.RegistrantID,
			&order.OrderNumber,
			&order.Amount,
			&order.Currency,
			&order.PaymentType,
			&order.PaymentGateway,
			&order.PaymentMethod,
			&order.PaymentChannel,
			&order.PaymentStatus,
			&order.PaymentToken,
			&order.PaymentURL,
			&order.PaymentTransactionID,
			&order.PaymentMetadata,
			&order.PaymentProofURL,
			&order.PaymentProofFilename,
			&order.VerifiedBy,
			&order.VerifiedAt,
			&order.PaymentTime,
			&order.ExpiresAt,
			&order.DaoEntity.Deleted,
			&order.DaoEntity.DataHash,
			&order.DaoEntity.CreatedAt,
			&order.DaoEntity.UpdatedAt,
		); err != nil {
			d.log.Error(ctx, "orderDAO.Search.Scan", zap.Error(err))
			return nil, err
		}

		orders = append(orders, order)
	}

	return orders, nil
}

func (d orderDAO) SearchForUpdate(ctx context.Context, query entity.OrderQuery) (entity.Orders, error) {

	sqlSelect := sqlgo.NewSQLGoSelect().
		SetSQLSelect("o.id", "id").
		SetSQLSelect("o.event_id", "event_id").
		SetSQLSelect("o.registrant_id", "registrant_id").
		SetSQLSelect("o.order_number", "order_number").
		SetSQLSelect("o.amount", "amount").
		SetSQLSelect("o.currency", "currency").
		SetSQLSelect("o.payment_type", "payment_type").
		SetSQLSelect("o.payment_gateway", "payment_gateway").
		SetSQLSelect("o.payment_method", "payment_method").
		SetSQLSelect("o.payment_channel", "payment_channel").
		SetSQLSelect("o.payment_status", "payment_status").
		SetSQLSelect("o.payment_token", "payment_token").
		SetSQLSelect("o.payment_url", "payment_url").
		SetSQLSelect("o.payment_transaction_id", "payment_transaction_id").
		SetSQLSelect("o.payment_metadata", "payment_metadata").
		SetSQLSelect("o.payment_proof_url", "payment_proof_url").
		SetSQLSelect("o.payment_proof_filename", "payment_proof_filename").
		SetSQLSelect("o.verified_by", "verified_by").
		SetSQLSelect("o.verified_at", "verified_at").
		SetSQLSelect("o.payment_time", "payment_time").
		SetSQLSelect("o.expires_at", "expires_at").
		SetSQLSelect("o.deleted", "deleted").
		SetSQLSelect("o.data_hash", "data_hash").
		SetSQLSelect("o.created_at", "created_at").
		SetSQLSelect("o.updated_at", "updated_at")

	sqlFrom := sqlgo.NewSQLGoFrom().
		SetSQLFrom("orders", "o")

	sqlWhere := sqlgo.NewSQLGoWhere()
	sqlWhere.SetSQLWhere("AND", "o.deleted", "=", false)

	if len(query.IDs) > 0 {
		sqlWhere.SetSQLWhere("AND", "o.id", "IN", query.IDs)
	}

	if len(query.RegistrantIDs) > 0 {
		sqlWhere.SetSQLWhere("AND", "o.registrant_id", "IN", query.RegistrantIDs)
	}

	if len(query.EventIDs) > 0 {
		sqlWhere.SetSQLWhere("AND", "o.event_id", "IN", query.EventIDs)
	}

	if len(query.OrderNumbers) > 0 {
		sqlWhere.SetSQLWhere("AND", "o.order_number", "IN", query.OrderNumbers)
	}

	if len(query.Statuses) > 0 {
		sqlWhere.SetSQLWhere("AND", "o.payment_status", "IN", query.Statuses)
	}

	if len(query.PaymentGateways) > 0 {
		sqlWhere.SetSQLWhere("AND", "o.payment_gateway", "IN", query.PaymentGateways)
	}

	sql := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLGoSelect(sqlSelect).
		SetSQLGoFrom(sqlFrom).
		SetSQLGoWhere(sqlWhere)

	sqlStr := sql.BuildSQL() + " FOR UPDATE"
	sqlParams := sql.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "orderDAO.SearchForUpdate",
		zap.String("SQL", sqlStr),
		zap.Any("Params", sqlParams),
	)

	rows, err := d.dbTrx.GetSqlTx().QueryContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		d.log.Error(ctx, "orderDAO.SearchForUpdate",
			zap.String("SQL", sqlStr),
			zap.Any("Params", sqlParams),
			zap.Error(err),
		)
		return nil, err
	}
	defer rows.Close()

	var orders entity.Orders
	for rows.Next() {
		var order entity.Order

		if err := rows.Scan(
			&order.ID,
			&order.EventID,
			&order.RegistrantID,
			&order.OrderNumber,
			&order.Amount,
			&order.Currency,
			&order.PaymentType,
			&order.PaymentGateway,
			&order.PaymentMethod,
			&order.PaymentChannel,
			&order.PaymentStatus,
			&order.PaymentToken,
			&order.PaymentURL,
			&order.PaymentTransactionID,
			&order.PaymentMetadata,
			&order.PaymentProofURL,
			&order.PaymentProofFilename,
			&order.VerifiedBy,
			&order.VerifiedAt,
			&order.PaymentTime,
			&order.ExpiresAt,
			&order.DaoEntity.Deleted,
			&order.DaoEntity.DataHash,
			&order.DaoEntity.CreatedAt,
			&order.DaoEntity.UpdatedAt,
		); err != nil {
			d.log.Error(ctx, "orderDAO.SearchForUpdate.Scan", zap.Error(err))
			return nil, err
		}

		orders = append(orders, order)
	}

	return orders, nil
}

func (d orderDAO) Insert(ctx context.Context, orders entity.Orders) error {
	if len(orders) < 1 {
		return fmt.Errorf("empty order data")
	}

	sqlInsert := sqlgo.NewSQLGoInsert().
		SetSQLInsert("orders").
		SetSQLInsertColumn(
			"id", "event_id", "registrant_id", "order_number", "amount", "currency",
			"payment_type", "payment_gateway", "payment_status", "payment_token", "payment_url",
			"payment_proof_url", "payment_proof_filename", "verified_by", "verified_at",
			"expires_at", "deleted", "data_hash", "created_at",
		)

	for i, order := range orders {
		order.CreatedAt = time.Now()

		if order.ID == "" {
			order.ID = pubEntity.MakeUUID(
				order.OrderNumber,
				string(order.RegistrantID),
				order.CreatedAt.String(),
			)
		}

		sqlInsert.SetSQLInsertValue(
			order.ID,
			order.EventID,
			order.RegistrantID,
			order.OrderNumber,
			order.Amount,
			order.Currency,
			order.PaymentType,
			order.PaymentGateway,
			order.PaymentStatus,
			order.PaymentToken,
			order.PaymentURL,
			order.PaymentProofURL,
			order.PaymentProofFilename,
			order.VerifiedBy,
			order.VerifiedAt,
			order.ExpiresAt,
			order.DaoEntity.Deleted,
			order.DaoEntity.DataHash,
			order.CreatedAt,
		)

		orders[i] = order
	}

	sql := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLGoInsert(sqlInsert)

	sqlStr := sql.BuildSQL()
	sqlParams := sql.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "orderDAO.Insert",
		zap.String("SQL", sqlStr),
		zap.Any("Params", sqlParams),
	)

	_, err := d.dbTrx.GetSqlTx().ExecContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		d.log.Error(ctx, "orderDAO.Insert",
			zap.String("SQL", sqlStr),
			zap.Any("Params", sqlParams),
			zap.Error(err),
		)
		return err
	}

	return nil
}

func (d orderDAO) Update(ctx context.Context, orders entity.Orders) error {
	if len(orders) < 1 {
		return fmt.Errorf("empty order data")
	}

	for i, order := range orders {
		now := time.Now()
		order.UpdatedAt = &now

		sql := sqlgo.NewSQLGo().
			SetSQLSchema("public").
			SetSQLUpdate("orders").
			SetSQLUpdateValue("event_id", order.EventID).
			SetSQLUpdateValue("amount", order.Amount).
			SetSQLUpdateValue("currency", order.Currency).
			SetSQLUpdateValue("payment_status", order.PaymentStatus).
			SetSQLUpdateValue("data_hash", order.DataHash).
			SetSQLUpdateValue("updated_at", order.UpdatedAt)

		if order.PaymentType != nil {
			sql.SetSQLUpdateValue("payment_type", order.PaymentType)
		}
		if order.PaymentGateway != nil {
			sql.SetSQLUpdateValue("payment_gateway", order.PaymentGateway)
		}
		if order.PaymentMethod != nil {
			sql.SetSQLUpdateValue("payment_method", order.PaymentMethod)
		}
		if order.PaymentChannel != nil {
			sql.SetSQLUpdateValue("payment_channel", order.PaymentChannel)
		}
		if order.PaymentToken != nil {
			sql.SetSQLUpdateValue("payment_token", order.PaymentToken)
		}
		if order.PaymentURL != nil {
			sql.SetSQLUpdateValue("payment_url", order.PaymentURL)
		}
		if order.PaymentTransactionID != nil {
			sql.SetSQLUpdateValue("payment_transaction_id", order.PaymentTransactionID)
		}
		if order.PaymentMetadata != nil {
			sql.SetSQLUpdateValue("payment_metadata", order.PaymentMetadata)
		}
		if order.PaymentProofURL != nil {
			sql.SetSQLUpdateValue("payment_proof_url", order.PaymentProofURL)
		}
		if order.PaymentProofFilename != nil {
			sql.SetSQLUpdateValue("payment_proof_filename", order.PaymentProofFilename)
		}
		if order.VerifiedBy != nil {
			sql.SetSQLUpdateValue("verified_by", order.VerifiedBy)
		}
		if order.VerifiedAt != nil {
			sql.SetSQLUpdateValue("verified_at", order.VerifiedAt)
		}
		if order.PaymentTime != nil {
			sql.SetSQLUpdateValue("payment_time", order.PaymentTime)
		}
		if order.ExpiresAt != nil {
			sql.SetSQLUpdateValue("expires_at", order.ExpiresAt)
		}

		sql.SetSQLWhere("AND", "id", "=", order.ID)

		sqlStr := sql.BuildSQL()
		sqlParams := sql.GetSQLGoParameter().GetSQLParameter()

		d.log.Debug(ctx, "orderDAO.Update",
			zap.String("SQL", sqlStr),
			zap.Any("Params", sqlParams),
		)

		_, err := d.dbTrx.GetSqlTx().ExecContext(ctx, sqlStr, sqlParams...)
		if err != nil {
			d.log.Error(ctx, "orderDAO.Update",
				zap.String("SQL", sqlStr),
				zap.Any("Params", sqlParams),
				zap.Error(err),
			)
			return err
		}

		orders[i] = order
	}

	return nil
}

func (d orderDAO) Delete(ctx context.Context, id pubEntity.UUID) error {
	sql := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLDelete("orders").
		SetSQLWhere("AND", "id", "=", id)

	sqlStr := sql.BuildSQL()
	sqlParams := sql.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "orderDAO.Delete",
		zap.String("SQL", sqlStr),
		zap.Any("Params", sqlParams),
	)

	_, err := d.dbTrx.GetSqlTx().ExecContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		d.log.Error(ctx, "orderDAO.Delete",
			zap.String("SQL", sqlStr),
			zap.Any("Params", sqlParams),
			zap.Error(err),
		)
		return err
	}

	return nil
}

func (d orderDAO) SoftDelete(ctx context.Context, id pubEntity.UUID) error {
	sql := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLUpdate("orders").
		SetSQLUpdateValue("deleted", true).
		SetSQLWhere("AND", "id", "=", id)

	sqlStr := sql.BuildSQL()
	sqlParams := sql.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "orderDAO.SoftDelete",
		zap.String("SQL", sqlStr),
		zap.Any("Params", sqlParams),
	)

	_, err := d.dbTrx.GetSqlTx().ExecContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		d.log.Error(ctx, "orderDAO.SoftDelete",
			zap.String("SQL", sqlStr),
			zap.Any("Params", sqlParams),
			zap.Error(err),
		)
		return err
	}

	return nil
}
