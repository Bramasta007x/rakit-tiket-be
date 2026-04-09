package dao

import (
	"context"
	"fmt"
	"time"

	baseDao "rakit-tiket-be/internal/pkg/dao"
	pubEntity "rakit-tiket-be/pkg/entity"
	appPayment "rakit-tiket-be/pkg/entity/app_payment"
	"rakit-tiket-be/pkg/util"

	"gitlab.com/threetopia/sqlgo/v2"
	"go.uber.org/zap"
)

type ManualTransferDAO interface {
	Search(ctx context.Context, query appPayment.ManualTransferQuery) (appPayment.ManualTransfers, error)
	GetByID(ctx context.Context, id pubEntity.UUID) (*appPayment.ManualTransfer, error)
	GetByOrderID(ctx context.Context, orderID pubEntity.UUID) (*appPayment.ManualTransfer, error)
	Insert(ctx context.Context, transfers appPayment.ManualTransfers) error
	Update(ctx context.Context, transfers appPayment.ManualTransfers) error
	Delete(ctx context.Context, id pubEntity.UUID) error
}

type manualTransferDAO struct {
	log   util.LogUtil
	dbTrx baseDao.DBTransaction
}

func MakeManualTransferDAO(log util.LogUtil, dbTrx baseDao.DBTransaction) ManualTransferDAO {
	return &manualTransferDAO{
		log:   log,
		dbTrx: dbTrx,
	}
}

func (d *manualTransferDAO) Search(ctx context.Context, query appPayment.ManualTransferQuery) (appPayment.ManualTransfers, error) {
	sqlSelect := sqlgo.NewSQLGoSelect().
		SetSQLSelect("mt.id", "id").
		SetSQLSelect("mt.order_id", "order_id").
		SetSQLSelect("mt.bank_account_id", "bank_account_id").
		SetSQLSelect("mt.transfer_amount", "transfer_amount").
		SetSQLSelect("mt.transfer_proof_url", "transfer_proof_url").
		SetSQLSelect("mt.transfer_proof_filename", "transfer_proof_filename").
		SetSQLSelect("mt.sender_name", "sender_name").
		SetSQLSelect("mt.sender_account_number", "sender_account_number").
		SetSQLSelect("mt.transfer_date", "transfer_date").
		SetSQLSelect("mt.admin_notes", "admin_notes").
		SetSQLSelect("mt.reviewed_by", "reviewed_by").
		SetSQLSelect("mt.reviewed_at", "reviewed_at").
		SetSQLSelect("mt.status", "status").
		SetSQLSelect("mt.deleted", "deleted").
		SetSQLSelect("mt.data_hash", "data_hash").
		SetSQLSelect("mt.created_at", "created_at").
		SetSQLSelect("mt.updated_at", "updated_at")

	sqlFrom := sqlgo.NewSQLGoFrom().
		SetSQLFrom("manual_transfers", "mt")

	sqlWhere := sqlgo.NewSQLGoWhere()
	sqlWhere.SetSQLWhere("AND", "mt.deleted", "=", false)

	if len(query.IDs) > 0 {
		sqlWhere.SetSQLWhere("AND", "mt.id", "IN", query.IDs)
	}

	if len(query.OrderIDs) > 0 {
		sqlWhere.SetSQLWhere("AND", "mt.order_id", "IN", query.OrderIDs)
	}

	if len(query.Statuses) > 0 {
		sqlWhere.SetSQLWhere("AND", "mt.status", "IN", query.Statuses)
	}

	sqlStmt := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLGoSelect(sqlSelect).
		SetSQLGoFrom(sqlFrom).
		SetSQLGoWhere(sqlWhere)

	sqlStr := sqlStmt.BuildSQL()
	sqlParams := sqlStmt.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "manualTransferDAO.Search",
		zap.String("SQL", sqlStr),
		zap.Any("Params", sqlParams),
	)

	rows, err := d.dbTrx.GetSqlDB().QueryContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		d.log.Error(ctx, "manualTransferDAO.Search", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var transfers appPayment.ManualTransfers
	for rows.Next() {
		var transfer appPayment.ManualTransfer
		if err := rows.Scan(
			&transfer.ID, &transfer.OrderID, &transfer.BankAccountID,
			&transfer.TransferAmount, &transfer.TransferProofURL,
			&transfer.TransferProofFilename, &transfer.SenderName,
			&transfer.SenderAccountNumber, &transfer.TransferDate,
			&transfer.AdminNotes, &transfer.ReviewedBy, &transfer.ReviewedAt,
			&transfer.Status, &transfer.Deleted, &transfer.DataHash,
			&transfer.CreatedAt, &transfer.UpdatedAt,
		); err != nil {
			d.log.Error(ctx, "manualTransferDAO.Search.Scan", zap.Error(err))
			return nil, err
		}
		transfers = append(transfers, transfer)
	}

	return transfers, nil
}

func (d *manualTransferDAO) GetByID(ctx context.Context, id pubEntity.UUID) (*appPayment.ManualTransfer, error) {
	transfers, err := d.Search(ctx, appPayment.ManualTransferQuery{IDs: []string{string(id)}})
	if err != nil {
		return nil, err
	}
	if len(transfers) == 0 {
		return nil, nil
	}
	return &transfers[0], nil
}

func (d *manualTransferDAO) GetByOrderID(ctx context.Context, orderID pubEntity.UUID) (*appPayment.ManualTransfer, error) {
	transfers, err := d.Search(ctx, appPayment.ManualTransferQuery{OrderIDs: []string{string(orderID)}})
	if err != nil {
		return nil, err
	}
	if len(transfers) == 0 {
		return nil, nil
	}
	return &transfers[0], nil
}

func (d *manualTransferDAO) Insert(ctx context.Context, transfers appPayment.ManualTransfers) error {
	if len(transfers) < 1 {
		return fmt.Errorf("empty manual transfer data")
	}

	sqlInsert := sqlgo.NewSQLGoInsert().
		SetSQLInsert("manual_transfers").
		SetSQLInsertColumn(
			"id", "order_id", "bank_account_id", "transfer_amount",
			"transfer_proof_url", "transfer_proof_filename",
			"sender_name", "sender_account_number", "transfer_date",
			"status", "deleted", "data_hash", "created_at",
		)

	for i, transfer := range transfers {
		now := time.Now()
		transfer.CreatedAt = now

		if transfer.ID == "" {
			transfer.ID = pubEntity.MakeUUID("MT", string(transfer.OrderID), now.String())
		}

		sqlInsert.SetSQLInsertValue(
			transfer.ID,
			transfer.OrderID,
			transfer.BankAccountID,
			transfer.TransferAmount,
			transfer.TransferProofURL,
			transfer.TransferProofFilename,
			transfer.SenderName,
			transfer.SenderAccountNumber,
			transfer.TransferDate,
			transfer.Status,
			transfer.Deleted,
			transfer.DataHash,
			transfer.CreatedAt,
		)

		transfers[i] = transfer
	}

	sqlStmt := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLGoInsert(sqlInsert)

	sqlStr := sqlStmt.BuildSQL()
	sqlParams := sqlStmt.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "manualTransferDAO.Insert",
		zap.String("SQL", sqlStr),
		zap.Any("Params", sqlParams),
	)

	_, err := d.dbTrx.GetSqlTx().ExecContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		d.log.Error(ctx, "manualTransferDAO.Insert", zap.Error(err))
		return err
	}

	return nil
}

func (d *manualTransferDAO) Update(ctx context.Context, transfers appPayment.ManualTransfers) error {
	if len(transfers) < 1 {
		return fmt.Errorf("empty manual transfer data")
	}

	for i, transfer := range transfers {
		now := time.Now()
		transfer.UpdatedAt = &now

		sqlStmt := sqlgo.NewSQLGo().
			SetSQLSchema("public").
			SetSQLUpdate("manual_transfers").
			SetSQLUpdateValue("transfer_amount", transfer.TransferAmount).
			SetSQLUpdateValue("transfer_proof_url", transfer.TransferProofURL).
			SetSQLUpdateValue("admin_notes", transfer.AdminNotes).
			SetSQLUpdateValue("reviewed_by", transfer.ReviewedBy).
			SetSQLUpdateValue("reviewed_at", transfer.ReviewedAt).
			SetSQLUpdateValue("status", transfer.Status).
			SetSQLUpdateValue("data_hash", transfer.DataHash).
			SetSQLUpdateValue("updated_at", transfer.UpdatedAt).
			SetSQLWhere("AND", "id", "=", transfer.ID)

		sqlStr := sqlStmt.BuildSQL()
		sqlParams := sqlStmt.GetSQLGoParameter().GetSQLParameter()

		d.log.Debug(ctx, "manualTransferDAO.Update",
			zap.String("SQL", sqlStr),
			zap.Any("Params", sqlParams),
		)

		_, err := d.dbTrx.GetSqlTx().ExecContext(ctx, sqlStr, sqlParams...)
		if err != nil {
			d.log.Error(ctx, "manualTransferDAO.Update", zap.Error(err))
			return err
		}

		transfers[i] = transfer
	}

	return nil
}

func (d *manualTransferDAO) Delete(ctx context.Context, id pubEntity.UUID) error {
	now := time.Now()

	sqlStmt := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLUpdate("manual_transfers").
		SetSQLUpdateValue("deleted", true).
		SetSQLUpdateValue("updated_at", &now).
		SetSQLWhere("AND", "id", "=", id)

	sqlStr := sqlStmt.BuildSQL()
	sqlParams := sqlStmt.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "manualTransferDAO.Delete",
		zap.String("SQL", sqlStr),
		zap.Any("Params", sqlParams),
	)

	_, err := d.dbTrx.GetSqlTx().ExecContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		d.log.Error(ctx, "manualTransferDAO.Delete", zap.Error(err))
		return err
	}

	return nil
}
