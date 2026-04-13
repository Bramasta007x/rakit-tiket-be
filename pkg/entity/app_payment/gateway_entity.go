package app_payment

import (
	pubEntity "rakit-tiket-be/pkg/entity"
)

type PaymentGateway struct {
	ID           pubEntity.UUID `json:"id"`
	Code         string         `json:"code"`
	Name         string         `json:"name"`
	IsEnabled    bool           `json:"is_enabled"`
	IsActive     bool           `json:"is_active"`
	DisplayOrder int            `json:"display_order"`
	pubEntity.DaoEntity
}

type PaymentGateways []PaymentGateway

type GatewayQuery struct {
	IDs       []string `query:"id"`
	Codes     []string `query:"code"`
	Names     []string `query:"name"`
	IsEnabled *bool    `query:"is_enabled"`
	IsActive  *bool    `query:"is_active"`
}

type PaymentSetting struct {
	ID           pubEntity.UUID `json:"id"`
	SettingKey   string         `json:"setting_key"`
	SettingValue bool           `json:"setting_value"`
	DisplayOrder int            `json:"display_order"`
	pubEntity.DaoEntity
}

type PaymentSettings []PaymentSetting

type PaymentSettingQuery struct {
	SettingKeys []string `query:"setting_key"`
}

const (
	SettingKeyManualTransferEnabled = "MANUAL_TRANSFER_ENABLED"
)
