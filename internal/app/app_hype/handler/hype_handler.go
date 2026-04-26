package handler

import (
	"net/http"

	"rakit-tiket-be/internal/app/app_hype/service"
	"rakit-tiket-be/internal/pkg/middleware"
	"rakit-tiket-be/pkg/util"

	"github.com/labstack/echo/v4"
)

type HypeHandler interface {
	RegisterRouter(g *echo.Group)
}

type hypeHandler struct {
	log            util.LogUtil
	hypeService    service.HypeService
	authMiddleware middleware.AuthMiddleware
}

func MakeHypeHandler(log util.LogUtil, hypeService service.HypeService, authMiddleware middleware.AuthMiddleware) HypeHandler {
	return &hypeHandler{
		log:            log,
		hypeService:    hypeService,
		authMiddleware: authMiddleware,
	}
}

func (h *hypeHandler) RegisterRouter(g *echo.Group) {
	public := g.Group("/v1")
	public.GET("/events/:event_id/hype", h.getActiveTickets)
	public.GET("/hype/check/:ticket_id", h.checkAvailability)

	admin := g.Group("/v1/admin")
	admin.Use(h.authMiddleware.VerifyToken)
	admin.Use(h.authMiddleware.RequireAdmin)

	admin.POST("/hype/flash-sale", h.setFlashSale)
	admin.DELETE("/hype/flash-sale/:ticket_id", h.disableFlashSale)
	admin.PUT("/hype/countdown", h.setCountdown)
	admin.PUT("/hype/stock-alert", h.setStockAlert)
}

func (h *hypeHandler) getActiveTickets(c echo.Context) error {
	eventID := c.Param("event_id")
	if eventID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "event_id is required")
	}

	tickets, err := h.hypeService.GetActiveTickets(c.Request().Context(), eventID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    tickets,
		"count":   len(tickets),
	})
}

func (h *hypeHandler) checkAvailability(c echo.Context) error {
	ticketID := c.Param("ticket_id")
	if ticketID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "ticket_id is required")
	}

	check, err := h.hypeService.CheckAvailability(c.Request().Context(), ticketID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    check,
	})
}

func (h *hypeHandler) setFlashSale(c echo.Context) error {
	var req service.SetFlashSaleRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if req.TicketID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "ticket_id is required")
	}

	if req.FlashPrice <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "flash_price must be greater than 0")
	}

	if err := h.hypeService.SetFlashSale(c.Request().Context(), req); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Flash sale berhasil diaktifkan",
	})
}

func (h *hypeHandler) disableFlashSale(c echo.Context) error {
	ticketID := c.Param("ticket_id")
	if ticketID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "ticket_id is required")
	}

	if err := h.hypeService.DisableFlashSale(c.Request().Context(), ticketID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Flash sale berhasil dinonaktifkan",
	})
}

func (h *hypeHandler) setCountdown(c echo.Context) error {
	var req service.SetCountdownRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if req.TicketID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "ticket_id is required")
	}

	if err := h.hypeService.SetCountdown(c.Request().Context(), req); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Countdown berhasil disetting",
	})
}

func (h *hypeHandler) setStockAlert(c echo.Context) error {
	var req service.SetStockAlertRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if req.TicketID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "ticket_id is required")
	}

	if err := h.hypeService.SetStockAlert(c.Request().Context(), req); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Stock alert berhasil disetting",
	})
}
