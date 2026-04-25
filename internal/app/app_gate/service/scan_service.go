package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"rakit-tiket-be/internal/app/app_gate/dao"
	pubEntity "rakit-tiket-be/pkg/entity"
	app_gate_config "rakit-tiket-be/pkg/entity/app_gate_config"
	app_gate_log "rakit-tiket-be/pkg/entity/app_gate_log"
	app_physical_ticket "rakit-tiket-be/pkg/entity/app_physical_ticket"
	"rakit-tiket-be/pkg/util"

	"go.uber.org/zap"
)

type ScanService interface {
	ScanTicket(ctx context.Context, qrCode, gateName, scannedBy string) (*ScanResult, error)
	GetGateLogs(ctx context.Context, eventID string, limit int) ([]GateLogInfo, error)
}

type ScanResult struct {
	Success    bool   `json:"success"`
	Action     string `json:"action"`
	QRCode     string `json:"qr_code"`
	TicketType string `json:"ticket_type"`
	ScanCount  int    `json:"scan_count"`
	Message    string `json:"message"`
}

type GateLogInfo struct {
	ID               string `json:"id"`
	EventID          string `json:"event_id"`
	PhysicalTicketID string `json:"physical_ticket_id"`
	Action           string `json:"action"`
	Success          bool   `json:"success"`
	Message          string `json:"message"`
	GateName         string `json:"gate_name"`
	TicketType       string `json:"ticket_type"`
	ScanSequence     int    `json:"scan_sequence"`
	CreatedAt        string `json:"created_at"`
}

type scanService struct {
	log   util.LogUtil
	sqlDB *sql.DB
}

func MakeScanService(log util.LogUtil, sqlDB *sql.DB) ScanService {
	return &scanService{
		log:   log,
		sqlDB: sqlDB,
	}
}

func (s *scanService) ScanTicket(ctx context.Context, qrCode, gateName, scannedBy string) (*ScanResult, error) {
	dbTrx := dao.NewTransactionGate(ctx, s.log, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	ticket, err := dbTrx.GetPhysicalTicketDAO().SearchByQRCode(ctx, qrCode)
	if err != nil {
		return &ScanResult{
			Success: false,
			Action:  "INVALID",
			QRCode:  qrCode,
			Message: "Tiket tidak ditemukan",
		}, nil
	}

	if ticket.Status == app_physical_ticket.PhysicalTicketStatusVoid {
		return &ScanResult{
			Success: false,
			Action:  "INVALID",
			QRCode:  qrCode,
			Message: "Tiket sudah dibatalkan",
		}, nil
	}

	config, err := dbTrx.GetGateConfigDAO().GetByEventID(ctx, string(ticket.EventID))
	if err != nil {
		return nil, fmt.Errorf("konfigurasi gate tidak ditemukan: %v", err)
	}

	if !config.IsActive {
		return nil, fmt.Errorf("gate check-in tidak aktif")
	}

	maxScan := config.MaxScanPerTicket
	if config.MaxScanByType != nil {
		if override, ok := config.MaxScanByType[ticket.TicketType]; ok {
			maxScan = override
		}
	}

	if ticket.ScanCount >= maxScan {
		s.log.Info(ctx, "scan exceeded max", zap.Int("scan_count", ticket.ScanCount), zap.Int("max_scan", maxScan))

		s.logAudit(ctx, dbTrx, ticket, "EXCEEDED", false, "Melebihi max scan", gateName, scannedBy)

		return &ScanResult{
			Success:    false,
			Action:     "EXCEEDED",
			QRCode:     qrCode,
			TicketType: ticket.TicketType,
			ScanCount:  ticket.ScanCount,
			Message:    fmt.Sprintf("Melebihi batas scan (%d/%d)", ticket.ScanCount, maxScan),
		}, nil
	}

	var action string
	var newStatus app_physical_ticket.PhysicalTicketStatus

	if config.Mode == app_gate_config.GateModeCheckInOut {
		if ticket.ScanCount == 0 {
			action = "CHECK_IN"
			newStatus = app_physical_ticket.PhysicalTicketStatusCheckedIn
			now := time.Now()
			ticket.CheckedInAt = &now
		} else {
			action = "CHECK_OUT"
			newStatus = app_physical_ticket.PhysicalTicketStatusCheckedOut
			now := time.Now()
			ticket.CheckedOutAt = &now
		}
	} else {
		action = "CHECK_IN"
		newStatus = app_physical_ticket.PhysicalTicketStatusCheckedIn
		now := time.Now()
		ticket.CheckedInAt = &now
	}

	ticket.ScanCount++
	ticket.Status = newStatus

	if err := dbTrx.GetPhysicalTicketDAO().Update(ctx, app_physical_ticket.PhysicalTickets{*ticket}); err != nil {
		return nil, fmt.Errorf("gagal update ticket: %v", err)
	}

	s.logAudit(ctx, dbTrx, ticket, action, true, fmt.Sprintf("%s berhasil", action), gateName, scannedBy)

	if err := dbTrx.GetSqlTx().Commit(); err != nil {
		return nil, err
	}

	return &ScanResult{
		Success:    true,
		Action:     action,
		QRCode:     qrCode,
		TicketType: ticket.TicketType,
		ScanCount:  ticket.ScanCount,
		Message:    fmt.Sprintf("%s berhasil", action),
	}, nil
}

func (s *scanService) logAudit(ctx context.Context, dbTrx dao.DBTransaction, ticket *app_physical_ticket.PhysicalTicket, action string, success bool, message, gateName, scannedBy string) {
	now := time.Now()

	logEntry := app_gate_log.GateLog{
		ID:               pubEntity.MakeUUID(string(ticket.ID), action, now.String()),
		EventID:          ticket.EventID,
		PhysicalTicketID: ticket.ID,
		ScannedBy:        scannedBy,
		Action:           app_gate_log.GateLogAction(action),
		Success:          success,
		Message:          message,
		GateName:         gateName,
		TicketType:       ticket.TicketType,
		ScanSequence:     ticket.ScanCount,
		CreatedAt:        now,
	}

	_ = dbTrx.GetGateLogDAO().Insert(ctx, logEntry)
}

func (s *scanService) GetGateLogs(ctx context.Context, eventID string, limit int) ([]GateLogInfo, error) {
	dbTrx := dao.NewTransactionGate(ctx, s.log, s.sqlDB)

	logs, err := dbTrx.GetGateLogDAO().Search(ctx, app_gate_log.GateLogQuery{
		EventIDs: []string{eventID},
	})
	if err != nil {
		return nil, err
	}

	if limit > 0 && len(logs) > limit {
		logs = logs[:limit]
	}

	var result []GateLogInfo
	for _, l := range logs {
		result = append(result, GateLogInfo{
			ID:               string(l.ID),
			EventID:          string(l.EventID),
			PhysicalTicketID: string(l.PhysicalTicketID),
			Action:           string(l.Action),
			Success:          l.Success,
			Message:          l.Message,
			GateName:         l.GateName,
			TicketType:       l.TicketType,
			ScanSequence:     l.ScanSequence,
			CreatedAt:        l.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return result, nil
}
