package handler

import (
	"net/http"

	"rakit-tiket-be/internal/app/app_checkin/service"
	"rakit-tiket-be/internal/pkg/middleware"
	"rakit-tiket-be/pkg/util"

	"github.com/labstack/echo/v4"
)

type GateHandler interface {
	RegisterRouter(g *echo.Group)
}

type gateHandler struct {
	log            util.LogUtil
	gateService    service.GateService
	scanService    service.ScanService
	authMiddleware middleware.AuthMiddleware
}

func MakeGateHandler(
	log util.LogUtil,
	gateService service.GateService,
	scanService service.ScanService,
	authMiddleware middleware.AuthMiddleware,
) GateHandler {
	return &gateHandler{
		log:            log,
		gateService:    gateService,
		scanService:    scanService,
		authMiddleware: authMiddleware,
	}
}

func (h *gateHandler) RegisterRouter(g *echo.Group) {
	public := g.Group("/v1")

	public.POST("/gate/scan", h.scanTicket)

	admin := g.Group("/v1/admin")
	admin.Use(h.authMiddleware.VerifyToken)
	admin.Use(h.authMiddleware.RequireAdmin)

	admin.POST("/gate/config", h.createGateConfig)
	admin.GET("/gate/config/:event_id", h.getGateConfig)

	admin.POST("/gate/generate-qr", h.generatePhysicalTickets)
	admin.GET("/gate/qr/:event_id", h.getPhysicalTickets)

	admin.GET("/gate/stats/:event_id", h.getGateStats)
	admin.GET("/gate/logs/:event_id", h.getGateLogs)
}

func (h *gateHandler) createGateConfig(c echo.Context) error {
	var req service.CreateGateConfigRequest

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if req.EventID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "event_id is required")
	}

	if req.Mode == "" {
		req.Mode = "CHECK_IN"
	}

	if req.MaxScanPerTicket == 0 {
		req.MaxScanPerTicket = 1
	}

	data, err := h.gateService.CreateGateConfig(c.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

func (h *gateHandler) getGateConfig(c echo.Context) error {
	eventID := c.Param("event_id")

	if eventID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "event_id is required")
	}

	data, err := h.gateService.GetGateConfig(c.Request().Context(), eventID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "konfigurasi gate tidak ditemukan")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

func (h *gateHandler) generatePhysicalTickets(c echo.Context) error {
	var req service.GenerateTicketsRequest

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if req.EventID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "event_id is required")
	}

	data, err := h.gateService.GeneratePhysicalTickets(c.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

func (h *gateHandler) getPhysicalTickets(c echo.Context) error {
	eventID := c.Param("event_id")

	if eventID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "event_id is required")
	}

	ticketTypes := c.QueryParam("ticket_types")
	var types []string
	if ticketTypes != "" {
		types = []string{ticketTypes}
	}

	data, err := h.gateService.GetPhysicalTickets(c.Request().Context(), eventID, types)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    data,
		"count":   len(data),
	})
}

func (h *gateHandler) getGateStats(c echo.Context) error {
	eventID := c.Param("event_id")

	if eventID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "event_id is required")
	}

	data, err := h.gateService.GetGateStats(c.Request().Context(), eventID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

func (h *gateHandler) scanTicket(c echo.Context) error {
	var req struct {
		QRCode   string `json:"qr_code"`
		GateName string `json:"gate_name"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if req.QRCode == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "qr_code is required")
	}

	data, err := h.scanService.ScanTicket(c.Request().Context(), req.QRCode, req.GateName, "")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if !data.Success {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"data":    data,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

func (h *gateHandler) getGateLogs(c echo.Context) error {
	eventID := c.Param("event_id")

	if eventID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "event_id is required")
	}

	limit := 100
	data, err := h.scanService.GetGateLogs(c.Request().Context(), eventID, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    data,
		"count":   len(data),
	})
}
