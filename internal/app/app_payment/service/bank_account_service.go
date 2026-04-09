package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"rakit-tiket-be/internal/app/app_payment/dao"
	pubEntity "rakit-tiket-be/pkg/entity"
	appPayment "rakit-tiket-be/pkg/entity/app_payment"
	"rakit-tiket-be/pkg/util"
)

var (
	ErrBankAccountNotFound = errors.New("bank account not found")
)

type BankAccountService interface {
	GetActiveBankAccounts(ctx context.Context) (appPayment.BankAccounts, error)
	CreateBankAccount(ctx context.Context, req CreateBankAccountRequest) (*appPayment.BankAccount, error)
	UpdateBankAccount(ctx context.Context, req UpdateBankAccountRequest) error
	DeleteBankAccount(ctx context.Context, id string) error
}

type bankAccountService struct {
	log   util.LogUtil
	sqlDB *sql.DB
}

func MakeBankAccountService(log util.LogUtil, sqlDB *sql.DB) BankAccountService {
	return &bankAccountService{
		log:   log,
		sqlDB: sqlDB,
	}
}

func (s *bankAccountService) GetActiveBankAccounts(ctx context.Context) (appPayment.BankAccounts, error) {
	dbTrx := dao.NewTransactionPayment(ctx, s.log, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()
	return dbTrx.GetBankAccountDAO().Search(ctx)
}

func (s *bankAccountService) CreateBankAccount(ctx context.Context, req CreateBankAccountRequest) (*appPayment.BankAccount, error) {
	dbTrx := dao.NewTransactionPayment(ctx, s.log, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	now := time.Now()
	account := appPayment.BankAccount{
		ID:              pubEntity.MakeUUID("BANK", req.AccountNumber, now.String()),
		BankName:        req.BankName,
		BankCode:        req.BankCode,
		AccountNumber:   req.AccountNumber,
		AccountHolder:   req.AccountHolder,
		IsActive:        req.IsActive,
		IsDefault:       req.IsDefault,
		InstructionText: req.InstructionText,
		DaoEntity: pubEntity.DaoEntity{
			Deleted:   false,
			CreatedAt: now,
		},
	}

	if err := dbTrx.GetBankAccountDAO().Insert(ctx, appPayment.BankAccounts{account}); err != nil {
		return nil, err
	}

	if err := dbTrx.GetSqlTx().Commit(); err != nil {
		return nil, err
	}

	return &account, nil
}

func (s *bankAccountService) UpdateBankAccount(ctx context.Context, req UpdateBankAccountRequest) error {
	dbTrx := dao.NewTransactionPayment(ctx, s.log, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	account, err := dbTrx.GetBankAccountDAO().GetByID(ctx, pubEntity.UUID(req.ID))
	if err != nil {
		return err
	}
	if account == nil {
		return ErrBankAccountNotFound
	}

	account.BankName = req.BankName
	account.BankCode = req.BankCode
	account.AccountNumber = req.AccountNumber
	account.AccountHolder = req.AccountHolder
	account.IsActive = req.IsActive
	account.IsDefault = req.IsDefault
	account.InstructionText = req.InstructionText

	if err := dbTrx.GetBankAccountDAO().Update(ctx, appPayment.BankAccounts{*account}); err != nil {
		return err
	}

	return dbTrx.GetSqlTx().Commit()
}

func (s *bankAccountService) DeleteBankAccount(ctx context.Context, id string) error {
	dbTrx := dao.NewTransactionPayment(ctx, s.log, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	if err := dbTrx.GetBankAccountDAO().Delete(ctx, pubEntity.UUID(id)); err != nil {
		return err
	}

	return dbTrx.GetSqlTx().Commit()
}

type CreateBankAccountRequest struct {
	BankName        string  `json:"bank_name" validate:"required"`
	BankCode        string  `json:"bank_code" validate:"required"`
	AccountNumber   string  `json:"account_number" validate:"required"`
	AccountHolder   string  `json:"account_holder" validate:"required"`
	IsActive        bool    `json:"is_active"`
	IsDefault       bool    `json:"is_default"`
	InstructionText *string `json:"instruction_text"`
}

type UpdateBankAccountRequest struct {
	ID              string  `json:"id" validate:"required"`
	BankName        string  `json:"bank_name" validate:"required"`
	BankCode        string  `json:"bank_code" validate:"required"`
	AccountNumber   string  `json:"account_number" validate:"required"`
	AccountHolder   string  `json:"account_holder" validate:"required"`
	IsActive        bool    `json:"is_active"`
	IsDefault       bool    `json:"is_default"`
	InstructionText *string `json:"instruction_text"`
}
