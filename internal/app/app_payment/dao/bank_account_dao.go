package dao

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	baseDao "rakit-tiket-be/internal/pkg/dao"
	pubEntity "rakit-tiket-be/pkg/entity"
	appPayment "rakit-tiket-be/pkg/entity/app_payment"
	"rakit-tiket-be/pkg/util"

	"gitlab.com/threetopia/sqlgo/v2"
	"go.uber.org/zap"
)

type BankAccountDAO interface {
	Search(ctx context.Context) (appPayment.BankAccounts, error)
	GetByID(ctx context.Context, id pubEntity.UUID) (*appPayment.BankAccount, error)
	Insert(ctx context.Context, accounts appPayment.BankAccounts) error
	Update(ctx context.Context, accounts appPayment.BankAccounts) error
	Delete(ctx context.Context, id pubEntity.UUID) error
}

type bankAccountDAO struct {
	log   util.LogUtil
	dbTrx baseDao.DBTransaction
}

func MakeBankAccountDAO(log util.LogUtil, dbTrx baseDao.DBTransaction) BankAccountDAO {
	return &bankAccountDAO{
		log:   log,
		dbTrx: dbTrx,
	}
}

func (d *bankAccountDAO) Search(ctx context.Context) (appPayment.BankAccounts, error) {
	sqlSelect := sqlgo.NewSQLGoSelect().
		SetSQLSelect("id", "id").
		SetSQLSelect("bank_name", "bank_name").
		SetSQLSelect("bank_code", "bank_code").
		SetSQLSelect("account_number", "account_number").
		SetSQLSelect("account_holder", "account_holder").
		SetSQLSelect("is_active", "is_active").
		SetSQLSelect("is_default", "is_default").
		SetSQLSelect("instruction_text", "instruction_text").
		SetSQLSelect("deleted", "deleted").
		SetSQLSelect("data_hash", "data_hash").
		SetSQLSelect("created_at", "created_at").
		SetSQLSelect("updated_at", "updated_at")

	sqlFrom := sqlgo.NewSQLGoFrom().
		SetSQLFrom("bank_accounts", "ba")

	sqlWhere := sqlgo.NewSQLGoWhere()
	sqlWhere.SetSQLWhere("AND", "ba.deleted", "=", false)
	sqlWhere.SetSQLWhere("AND", "ba.is_active", "=", true)

	sqlStmt := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLGoSelect(sqlSelect).
		SetSQLGoFrom(sqlFrom).
		SetSQLGoWhere(sqlWhere)

	sqlStr := sqlStmt.BuildSQL()
	sqlParams := sqlStmt.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "bankAccountDAO.Search",
		zap.String("SQL", sqlStr),
		zap.Any("Params", sqlParams),
	)

	rows, err := d.dbTrx.GetSqlDB().QueryContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		d.log.Error(ctx, "bankAccountDAO.Search", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var accounts appPayment.BankAccounts
	for rows.Next() {
		var account appPayment.BankAccount
		if err := rows.Scan(
			&account.ID, &account.BankName, &account.BankCode,
			&account.AccountNumber, &account.AccountHolder,
			&account.IsActive, &account.IsDefault, &account.InstructionText,
			&account.Deleted, &account.DataHash,
			&account.CreatedAt, &account.UpdatedAt,
		); err != nil {
			d.log.Error(ctx, "bankAccountDAO.Search.Scan", zap.Error(err))
			return nil, err
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

func (d *bankAccountDAO) GetByID(ctx context.Context, id pubEntity.UUID) (*appPayment.BankAccount, error) {
	sqlSelect := sqlgo.NewSQLGoSelect().
		SetSQLSelect("id", "id").
		SetSQLSelect("bank_name", "bank_name").
		SetSQLSelect("bank_code", "bank_code").
		SetSQLSelect("account_number", "account_number").
		SetSQLSelect("account_holder", "account_holder").
		SetSQLSelect("is_active", "is_active").
		SetSQLSelect("is_default", "is_default").
		SetSQLSelect("instruction_text", "instruction_text").
		SetSQLSelect("deleted", "deleted").
		SetSQLSelect("data_hash", "data_hash").
		SetSQLSelect("created_at", "created_at").
		SetSQLSelect("updated_at", "updated_at")

	sqlFrom := sqlgo.NewSQLGoFrom().
		SetSQLFrom("bank_accounts", "ba")

	sqlWhere := sqlgo.NewSQLGoWhere()
	sqlWhere.SetSQLWhere("AND", "ba.deleted", "=", false)
	sqlWhere.SetSQLWhere("AND", "ba.id", "=", id)

	sqlStmt := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLGoSelect(sqlSelect).
		SetSQLGoFrom(sqlFrom).
		SetSQLGoWhere(sqlWhere)

	sqlStr := sqlStmt.BuildSQL()
	sqlParams := sqlStmt.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "bankAccountDAO.GetByID",
		zap.String("SQL", sqlStr),
		zap.Any("Params", sqlParams),
	)

	row := d.dbTrx.GetSqlDB().QueryRowContext(ctx, sqlStr, sqlParams...)

	var account appPayment.BankAccount
	if err := row.Scan(
		&account.ID, &account.BankName, &account.BankCode,
		&account.AccountNumber, &account.AccountHolder,
		&account.IsActive, &account.IsDefault, &account.InstructionText,
		&account.Deleted, &account.DataHash,
		&account.CreatedAt, &account.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		d.log.Error(ctx, "bankAccountDAO.GetByID.Scan", zap.Error(err))
		return nil, err
	}

	return &account, nil
}

func (d *bankAccountDAO) Insert(ctx context.Context, accounts appPayment.BankAccounts) error {
	if len(accounts) < 1 {
		return fmt.Errorf("empty bank account data")
	}

	sqlInsert := sqlgo.NewSQLGoInsert().
		SetSQLInsert("bank_accounts").
		SetSQLInsertColumn(
			"id", "bank_name", "bank_code", "account_number", "account_holder",
			"is_active", "is_default", "instruction_text",
			"deleted", "data_hash", "created_at",
		)

	for i, account := range accounts {
		now := time.Now()
		account.CreatedAt = now

		if account.ID == "" {
			account.ID = pubEntity.MakeUUID("BANK", account.AccountNumber, now.String())
		}

		sqlInsert.SetSQLInsertValue(
			account.ID,
			account.BankName,
			account.BankCode,
			account.AccountNumber,
			account.AccountHolder,
			account.IsActive,
			account.IsDefault,
			account.InstructionText,
			account.Deleted,
			account.DataHash,
			account.CreatedAt,
		)

		accounts[i] = account
	}

	sqlStmt := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLGoInsert(sqlInsert)

	sqlStr := sqlStmt.BuildSQL()
	sqlParams := sqlStmt.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "bankAccountDAO.Insert",
		zap.String("SQL", sqlStr),
		zap.Any("Params", sqlParams),
	)

	_, err := d.dbTrx.GetSqlTx().ExecContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		d.log.Error(ctx, "bankAccountDAO.Insert", zap.Error(err))
		return err
	}

	return nil
}

func (d *bankAccountDAO) Update(ctx context.Context, accounts appPayment.BankAccounts) error {
	if len(accounts) < 1 {
		return fmt.Errorf("empty bank account data")
	}

	for i, account := range accounts {
		now := time.Now()
		account.UpdatedAt = &now

		sqlStmt := sqlgo.NewSQLGo().
			SetSQLSchema("public").
			SetSQLUpdate("bank_accounts").
			SetSQLUpdateValue("bank_name", account.BankName).
			SetSQLUpdateValue("bank_code", account.BankCode).
			SetSQLUpdateValue("account_number", account.AccountNumber).
			SetSQLUpdateValue("account_holder", account.AccountHolder).
			SetSQLUpdateValue("is_active", account.IsActive).
			SetSQLUpdateValue("is_default", account.IsDefault).
			SetSQLUpdateValue("instruction_text", account.InstructionText).
			SetSQLUpdateValue("data_hash", account.DataHash).
			SetSQLUpdateValue("updated_at", account.UpdatedAt).
			SetSQLWhere("AND", "id", "=", account.ID)

		sqlStr := sqlStmt.BuildSQL()
		sqlParams := sqlStmt.GetSQLGoParameter().GetSQLParameter()

		d.log.Debug(ctx, "bankAccountDAO.Update",
			zap.String("SQL", sqlStr),
			zap.Any("Params", sqlParams),
		)

		_, err := d.dbTrx.GetSqlTx().ExecContext(ctx, sqlStr, sqlParams...)
		if err != nil {
			d.log.Error(ctx, "bankAccountDAO.Update", zap.Error(err))
			return err
		}

		accounts[i] = account
	}

	return nil
}

func (d *bankAccountDAO) Delete(ctx context.Context, id pubEntity.UUID) error {
	now := time.Now()

	sqlStmt := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLUpdate("bank_accounts").
		SetSQLUpdateValue("deleted", true).
		SetSQLUpdateValue("updated_at", &now).
		SetSQLWhere("AND", "id", "=", id)

	sqlStr := sqlStmt.BuildSQL()
	sqlParams := sqlStmt.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "bankAccountDAO.Delete",
		zap.String("SQL", sqlStr),
		zap.Any("Params", sqlParams),
	)

	_, err := d.dbTrx.GetSqlTx().ExecContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		d.log.Error(ctx, "bankAccountDAO.Delete", zap.Error(err))
		return err
	}

	return nil
}
