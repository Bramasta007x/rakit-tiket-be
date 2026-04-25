package app_gate_log

import (
	"time"

	pubEntity "rakit-tiket-be/pkg/entity"
)

type GateLogAction string

const (
	GateLogActionCheckIn   GateLogAction = "CHECK_IN"
	GateLogActionCheckOut  GateLogAction = "CHECK_OUT"
	GateLogActionInvalid   GateLogAction = "INVALID"
	GateLogActionDuplicate GateLogAction = "DUPLICATE"
	GateLogActionExceeded  GateLogAction = "EXCEEDED"
)

type GateLogQuery struct {
	IDs               []string `query:"id"`
	EventIDs          []string `query:"event_id"`
	PhysicalTicketIDs []string `query:"physical_ticket_id"`
	Actions           []string `query:"action"`
	GateNames         []string `query:"gate_name"`
}

type GateLog struct {
	ID      pubEntity.UUID `json:"id"`
	EventID pubEntity.UUID `json:"event_id"`

	PhysicalTicketID pubEntity.UUID `json:"physical_ticket_id"`
	ScannedBy        string         `json:"scanned_by"`

	Action  GateLogAction `json:"action"`
	Success bool          `json:"success"`
	Message string        `json:"message"`

	GateName     string `json:"gate_name"`
	TicketType   string `json:"ticket_type"`
	ScanSequence int    `json:"scan_sequence"`

	CreatedAt time.Time `json:"created_at"`
}

type GateLogs []GateLog
