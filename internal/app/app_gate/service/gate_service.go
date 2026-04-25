package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"rakit-tiket-be/internal/app/app_gate/dao"
	ticketDao "rakit-tiket-be/internal/app/app_ticket/dao"
	pubEntity "rakit-tiket-be/pkg/entity"
	app_gate_config "rakit-tiket-be/pkg/entity/app_gate_config"
	app_physical_ticket "rakit-tiket-be/pkg/entity/app_physical_ticket"
	ticketEntity "rakit-tiket-be/pkg/entity/app_ticket"
	"rakit-tiket-be/pkg/util"

	"go.uber.org/zap"
)

type GateService interface {
	GeneratePhysicalTickets(ctx context.Context, req GenerateTicketsRequest) (map[string]GenerateResult, error)
	GetPhysicalTickets(ctx context.Context, eventID string, ticketTypes []string) ([]PhysicalTicketInfo, error)
	GetGateConfig(ctx context.Context, eventID string) (*GateConfigInfo, error)
	CreateGateConfig(ctx context.Context, req CreateGateConfigRequest) (*GateConfigInfo, error)
	GetGateStats(ctx context.Context, eventID string) (*GateStats, error)
}

type GenerateResult struct {
	Generated int    `json:"generated"`
	StartCode string `json:"start_code"`
	EndCode   string `json:"end_code"`
}

type GenerateTicketsRequest struct {
	EventID     string         `json:"event_id"`
	TicketTypes []string       `json:"ticket_types"`
	Qty         map[string]int `json:"qty"`
}

type PhysicalTicketInfo struct {
	ID          string `json:"id"`
	QRCode      string `json:"qr_code"`
	QRCodeImage string `json:"qr_code_image"`
	TicketType  string `json:"ticket_type"`
	TicketID    string `json:"ticket_id"`
	Status      string `json:"status"`
	ScanCount   int    `json:"scan_count"`
}

type GateConfigInfo struct {
	ID               string         `json:"id"`
	EventID          string         `json:"event_id"`
	Mode             string         `json:"mode"`
	MaxScanPerTicket int            `json:"max_scan_per_ticket"`
	MaxScanByType    map[string]int `json:"max_scan_by_type"`
	IsActive         bool           `json:"is_active"`
}

type CreateGateConfigRequest struct {
	EventID          string         `json:"event_id"`
	Mode             string         `json:"mode"`
	MaxScanPerTicket int            `json:"max_scan_per_ticket"`
	MaxScanByType    map[string]int `json:"max_scan_by_type"`
	IsActive         bool           `json:"is_active"`
}

type GateStats struct {
	TotalPhysicalTickets int                  `json:"total_physical_tickets"`
	CheckedIn            int                  `json:"checked_in"`
	CheckedOut           int                  `json:"checked_out"`
	ActiveNow            int                  `json:"active_now"`
	ByType               map[string]TypeStats `json:"by_type"`
}

type TypeStats struct {
	Total      int `json:"total"`
	CheckedIn  int `json:"checked_in"`
	CheckedOut int `json:"checked_out"`
	ActiveNow  int `json:"active_now"`
}

type gateService struct {
	log   util.LogUtil
	sqlDB *sql.DB
}

func MakeGateService(log util.LogUtil, sqlDB *sql.DB) GateService {
	return &gateService{
		log:   log,
		sqlDB: sqlDB,
	}
}

func (s *gateService) GeneratePhysicalTickets(ctx context.Context, req GenerateTicketsRequest) (map[string]GenerateResult, error) {
	dbTrx := dao.NewTransactionGate(ctx, s.log, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	eventID := req.EventID

	ticketMasterMap, err := s.getTicketMasters(ctx, eventID, req.TicketTypes)
	if err != nil {
		return nil, err
	}

	result := make(map[string]GenerateResult)

	for ticketType, ticketMaster := range ticketMasterMap {
		requestedQty := 0
		if req.Qty != nil {
			if qty, ok := req.Qty[ticketType]; ok {
				requestedQty = qty
			}
		}

		if requestedQty <= 0 {
			continue
		}

		seq := 0

		for i := 0; i < requestedQty; i++ {
			qrCode := s.generateQRCode(ticketType, seq)
			seq++

			pt := pubEntity.MakeUUID(qrCode, time.Now().String())

			err := dbTrx.GetPhysicalTicketDAO().Insert(ctx, app_physical_ticket.PhysicalTickets{
				{
					ID:         pt,
					EventID:    pubEntity.UUID(eventID),
					TicketType: ticketType,
					TicketID:   ticketMaster.ID,
					QRCode:     qrCode,
					QRCodeHash: util.MakeUUID(qrCode),
				},
			})
			if err != nil {
				s.log.Error(ctx, "GeneratePhysicalTickets.Insert", zap.Error(err))
				return nil, err
			}
		}

		result[ticketType] = GenerateResult{
			Generated: requestedQty,
			StartCode: s.generateQRCode(ticketType, 0),
			EndCode:   s.generateQRCode(ticketType, requestedQty-1),
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("tidak ada ticket untuk digenerate")
	}

	if err := dbTrx.GetSqlTx().Commit(); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *gateService) generateQRCode(ticketType string, sequence int) string {
	year := time.Now().Year()
	prefix := fmt.Sprintf("TKT%d-%s", year, strings.ToUpper(ticketType))
	return fmt.Sprintf("%s-%d%03d", prefix, time.Now().UnixNano(), sequence+1)
}

func (s *gateService) getTicketMasters(ctx context.Context, eventID string, ticketTypes []string) (map[string]ticketEntity.Ticket, error) {
	ticketDB := ticketDao.NewTransactionTicket(ctx, s.log, s.sqlDB)

	query := ticketEntity.TicketQuery{
		EventIDs: []string{eventID},
	}

	if len(ticketTypes) > 0 {
		query.Types = ticketTypes
	}

	tickets, err := ticketDB.GetTicketDAO().Search(ctx, query)
	if err != nil {
		return nil, err
	}

	ticketMap := make(map[string]ticketEntity.Ticket)
	for _, t := range tickets {
		ticketMap[t.Type] = t
	}

	return ticketMap, nil
}

func (s *gateService) getExistingPhysicalTicketCount(ctx context.Context, eventID, ticketType string) (int, error) {
	dbTrx := dao.NewTransactionGate(ctx, s.log, s.sqlDB)

	tickets, err := dbTrx.GetPhysicalTicketDAO().Search(ctx, app_physical_ticket.PhysicalTicketQuery{
		EventIDs:    []string{eventID},
		TicketTypes: []string{ticketType},
	})
	if err != nil {
		return 0, err
	}

	return len(tickets), nil
}

func (s *gateService) GetPhysicalTickets(ctx context.Context, eventID string, ticketTypes []string) ([]PhysicalTicketInfo, error) {
	dbTrx := dao.NewTransactionGate(ctx, s.log, s.sqlDB)

	tickets, err := dbTrx.GetPhysicalTicketDAO().Search(ctx, app_physical_ticket.PhysicalTicketQuery{
		EventIDs:    []string{eventID},
		TicketTypes: ticketTypes,
	})
	if err != nil {
		return nil, err
	}

	var result []PhysicalTicketInfo
	for _, t := range tickets {
		qrCodeImage, _ := util.GenerateQRCodeBase64(t.QRCode, 256)
		result = append(result, PhysicalTicketInfo{
			ID:          string(t.ID),
			QRCode:      t.QRCode,
			QRCodeImage: qrCodeImage,
			TicketType:  t.TicketType,
			TicketID:    string(t.TicketID),
			Status:      string(t.Status),
			ScanCount:   t.ScanCount,
		})
	}

	return result, nil
}

func (s *gateService) GetGateConfig(ctx context.Context, eventID string) (*GateConfigInfo, error) {
	dbTrx := dao.NewTransactionGate(ctx, s.log, s.sqlDB)

	config, err := dbTrx.GetGateConfigDAO().GetByEventID(ctx, eventID)
	if err != nil {
		return nil, err
	}

	return &GateConfigInfo{
		ID:               string(config.ID),
		EventID:          string(config.EventID),
		Mode:             string(config.Mode),
		MaxScanPerTicket: config.MaxScanPerTicket,
		MaxScanByType:    config.MaxScanByType,
		IsActive:         config.IsActive,
	}, nil
}

func (s *gateService) CreateGateConfig(ctx context.Context, req CreateGateConfigRequest) (*GateConfigInfo, error) {
	dbTrx := dao.NewTransactionGate(ctx, s.log, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	config := app_gate_config.GateConfig{
		ID:               pubEntity.MakeUUID(req.EventID, time.Now().String()),
		EventID:          pubEntity.UUID(req.EventID),
		Mode:             app_gate_config.GateMode(req.Mode),
		MaxScanPerTicket: req.MaxScanPerTicket,
		MaxScanByType:    req.MaxScanByType,
		IsActive:         req.IsActive,
	}

	existing, _ := dbTrx.GetGateConfigDAO().GetByEventID(ctx, req.EventID)
	if existing != nil {
		existing.ID = config.ID
		existing.Mode = app_gate_config.GateMode(req.Mode)
		existing.MaxScanPerTicket = req.MaxScanPerTicket
		existing.MaxScanByType = req.MaxScanByType
		existing.IsActive = req.IsActive

		if err := dbTrx.GetGateConfigDAO().Update(ctx, *existing); err != nil {
			return nil, err
		}
	} else {
		if err := dbTrx.GetGateConfigDAO().Insert(ctx, config); err != nil {
			return nil, err
		}
	}

	if err := dbTrx.GetSqlTx().Commit(); err != nil {
		return nil, err
	}

	return &GateConfigInfo{
		ID:               string(config.ID),
		EventID:          string(config.EventID),
		Mode:             string(config.Mode),
		MaxScanPerTicket: config.MaxScanPerTicket,
		MaxScanByType:    config.MaxScanByType,
		IsActive:         config.IsActive,
	}, nil
}

func (s *gateService) GetGateStats(ctx context.Context, eventID string) (*GateStats, error) {
	dbTrx := dao.NewTransactionGate(ctx, s.log, s.sqlDB)

	tickets, err := dbTrx.GetPhysicalTicketDAO().Search(ctx, app_physical_ticket.PhysicalTicketQuery{
		EventIDs: []string{eventID},
	})
	if err != nil {
		return nil, err
	}

	stats := &GateStats{
		ByType: make(map[string]TypeStats),
	}

	for _, t := range tickets {
		stats.TotalPhysicalTickets++

		typeStats := stats.ByType[t.TicketType]
		typeStats.Total++

		if t.Status == app_physical_ticket.PhysicalTicketStatusCheckedIn {
			stats.CheckedIn++
			typeStats.CheckedIn++
		} else if t.Status == app_physical_ticket.PhysicalTicketStatusCheckedOut {
			stats.CheckedOut++
			typeStats.CheckedOut++
		}

		stats.ByType[t.TicketType] = typeStats
	}

	stats.ActiveNow = stats.CheckedIn

	for ticketType, ts := range stats.ByType {
		ts.ActiveNow = ts.CheckedIn
		stats.ByType[ticketType] = ts
	}

	return stats, nil
}
