package app_gate_config

import (
	pubEntity "rakit-tiket-be/pkg/entity"
)

type GateMode string

const (
	GateModeCheckIn    GateMode = "CHECK_IN"
	GateModeCheckInOut GateMode = "CHECK_IN_OUT"
)

type GateConfigQuery struct {
	IDs      []string `query:"id"`
	EventIDs []string `query:"event_id"`
	IsActive *bool    `query:"is_active"`
}

type GateConfig struct {
	ID      pubEntity.UUID `json:"id"`
	EventID pubEntity.UUID `json:"event_id"`

	Mode             GateMode `json:"mode"`
	MaxScanPerTicket int      `json:"max_scan_per_ticket"`

	MaxScanByType map[string]int `json:"max_scan_by_type"`

	IsActive bool `json:"is_active"`

	pubEntity.DaoEntity
}

type GateConfigs []GateConfig
