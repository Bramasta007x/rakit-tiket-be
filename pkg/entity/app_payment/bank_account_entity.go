package app_payment

import (
	pubEntity "rakit-tiket-be/pkg/entity"
)

type BankAccount struct {
	ID              pubEntity.UUID `json:"id"`
	BankName        string         `json:"bank_name"`
	BankCode        string         `json:"bank_code"`
	AccountNumber   string         `json:"account_number"`
	AccountHolder   string         `json:"account_holder"`
	IsActive        bool           `json:"is_active"`
	IsDefault       bool           `json:"is_default"`
	InstructionText *string        `json:"instruction_text"`
	pubEntity.DaoEntity
}

type BankAccounts []BankAccount
