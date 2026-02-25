package handler

import (
	"io"
	"net/http"
	"strings"

	"rakit-tiket-be/internal/app/app_order/service"
	"rakit-tiket-be/internal/pkg/payment"

	"github.com/labstack/echo/v4"
)

type OrderHandler interface {
	RegisterRouter(g *echo.Group)
}

type orderHandler struct {
	orderService service.OrderService
}

func MakeOrderHandler(orderService service.OrderService) OrderHandler {
	return orderHandler{
		orderService: orderService,
	}
}

func (h orderHandler) RegisterRouter(g *echo.Group) {
	public := g.Group("/v1")

	// Endpoint Webhook Payment Gateway
	// Contoh hit dari Midtrans: POST /v1/webhook/payment/midtrans
	public.POST("/webhook/payment/:gateway", h.handleWebhook)
}

func (h orderHandler) handleWebhook(c echo.Context) error {
	gatewayParam := strings.ToLower(c.Param("gateway"))
	var gateway payment.GatewayType

	// Validasi & Mapping param URL ke GatewayType Constant
	if gatewayParam == "midtrans" {
		gateway = payment.GatewayMidtrans
	} else if gatewayParam == "xendit" {
		gateway = payment.GatewayXendit
	} else {
		return echo.NewHTTPError(http.StatusBadRequest, "unsupported payment gateway")
	}

	// Baca body murni dari request (karena webhook butuh payload asli untuk validasi signature/parsing)
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to read webhook body")
	}
	defer c.Request().Body.Close()

	if err := h.orderService.HandleWebhook(c.Request().Context(), gateway, body); err != nil {
		// Log error di server, tapi dianjurkan tidak mengembalikan status error 500 ke gateway secara terus menerus
		// jika error karena validasi bisnis (seperti order tidak ditemukan).
		// Namun untuk mempermudah deteksi awal, kita lemparkan 500.
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Payment Gateway mengharapkan balasan HTTP 200 OK agar tidak melakukan retry berkali-kali
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Webhook processed successfully",
	})
}
