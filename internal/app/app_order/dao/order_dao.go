package dao

import (
	"context"
	"fmt"
	"time"

	baseDao "rakit-tiket-be/internal/pkg/dao"
	pubEntity "rakit-tiket-be/pkg/entity"
	entity "rakit-tiket-be/pkg/entity/app_order"

	"gitlab.com/threetopia/sqlgo/v2"
)

type OrderDAO interface {
	Search(ctx context.Context, query entity.OrderQuery) (entity.Orders, error)
	Insert(ctx context.Context, orders entity.Orders) error
	Update(ctx context.Context, orders entity.Orders) error
	Delete(ctx context.Context, id pubEntity.UUID) error
	SoftDelete(ctx context.Context, id pubEntity.UUID) error
}

func ordNullStr(s *string) interface{} {
	if s == nil {
		return nil
	}
	return *s
}

func ordNullTime(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return *t
}

type orderDAO struct {
	dbTrx baseDao.DBTransaction
}

func MakeOrderDAO(dbTrx baseDao.DBTransaction) OrderDAO {
	return orderDAO{
		dbTrx: dbTrx,
	}
}

func (d orderDAO) Search(ctx context.Context, query entity.OrderQuery) (entity.Orders, error) {

	sqlSelect := sqlgo.NewSQLGoSelect().
		SetSQLSelect("o.id", "id").
		SetSQLSelect("o.registrant_id", "registrant_id").
		SetSQLSelect("o.order_number", "order_number").
		SetSQLSelect("o.amount", "amount").
		SetSQLSelect("o.currency", "currency").
		SetSQLSelect("o.payment_gateway", "payment_gateway").
		SetSQLSelect("o.payment_method", "payment_method").
		SetSQLSelect("o.payment_channel", "payment_channel").
		SetSQLSelect("o.payment_status", "payment_status").
		SetSQLSelect("o.payment_token", "payment_token").
		SetSQLSelect("o.payment_url", "payment_url").
		SetSQLSelect("o.payment_transaction_id", "payment_transaction_id").
		SetSQLSelect("o.payment_metadata", "payment_metadata").
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

	rows, err := d.dbTrx.GetSqlDB().QueryContext(
		ctx,
		sql.BuildSQL(),
		sql.GetSQLGoParameter().GetSQLParameter()...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders entity.Orders
	for rows.Next() {
		var order entity.Order

		if err := rows.Scan(
			&order.ID,
			&order.RegistrantID,
			&order.OrderNumber,
			&order.Amount,
			&order.Currency,
			&order.PaymentGateway,
			&order.PaymentMethod,
			&order.PaymentChannel,
			&order.PaymentStatus,
			&order.PaymentToken,
			&order.PaymentURL,
			&order.PaymentTransactionID,
			&order.PaymentMetadata, // Ambil raw string/bytes JSON
			&order.PaymentTime,
			&order.ExpiresAt,
			&order.DaoEntity.Deleted,
			&order.DaoEntity.DataHash,
			&order.DaoEntity.CreatedAt,
			&order.DaoEntity.UpdatedAt,
		); err != nil {
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
			"id", "registrant_id", "order_number", "amount", "currency",
			"payment_gateway", "payment_method", "payment_channel", "payment_status",
			"payment_token", "payment_url", "payment_transaction_id", "payment_metadata",
			"payment_time", "expires_at", "deleted", "data_hash", "created_at",
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
			order.ID, order.RegistrantID, order.OrderNumber, order.Amount, order.Currency,
			ordNullStr(order.PaymentGateway), ordNullStr(order.PaymentMethod), ordNullStr(order.PaymentChannel), order.PaymentStatus,
			ordNullStr(order.PaymentToken), ordNullStr(order.PaymentURL), ordNullStr(order.PaymentTransactionID), ordNullStr(order.PaymentMetadata),
			ordNullTime(order.PaymentTime), ordNullTime(order.ExpiresAt), order.DaoEntity.Deleted, order.DaoEntity.DataHash, order.CreatedAt,
		)

		orders[i] = order
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
			SetSQLUpdateValue("amount", order.Amount).
			SetSQLUpdateValue("currency", order.Currency).
			SetSQLUpdateValue("payment_gateway", order.PaymentGateway).
			SetSQLUpdateValue("payment_method", order.PaymentMethod).
			SetSQLUpdateValue("payment_channel", order.PaymentChannel).
			SetSQLUpdateValue("payment_status", order.PaymentStatus).
			SetSQLUpdateValue("payment_token", order.PaymentToken).
			SetSQLUpdateValue("payment_url", order.PaymentURL).
			SetSQLUpdateValue("payment_transaction_id", order.PaymentTransactionID).
			SetSQLUpdateValue("payment_metadata", order.PaymentMetadata).
			SetSQLUpdateValue("payment_time", order.PaymentTime).
			SetSQLUpdateValue("expires_at", order.ExpiresAt).
			SetSQLUpdateValue("data_hash", order.DataHash).
			SetSQLUpdateValue("updated_at", order.UpdatedAt).
			SetSQLWhere("AND", "id", "=", order.ID)

		_, err := d.dbTrx.GetSqlTx().ExecContext(
			ctx,
			sql.BuildSQL(),
			sql.GetSQLGoParameter().GetSQLParameter()...,
		)
		if err != nil {
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

	_, err := d.dbTrx.GetSqlTx().ExecContext(
		ctx,
		sql.BuildSQL(),
		sql.GetSQLGoParameter().GetSQLParameter()...,
	)

	return err
}

func (d orderDAO) SoftDelete(ctx context.Context, id pubEntity.UUID) error {
	sql := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLUpdate("orders").
		SetSQLUpdateValue("deleted", true).
		SetSQLWhere("AND", "id", "=", id)

	_, err := d.dbTrx.GetSqlTx().ExecContext(
		ctx,
		sql.BuildSQL(),
		sql.GetSQLGoParameter().GetSQLParameter()...,
	)

	return err
}
