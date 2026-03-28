package handler

import (
	"io"
	"net/http"
	"strings"

	"rakit-tiket-be/internal/app/app_order/service"
	"rakit-tiket-be/internal/pkg/payment"
	"rakit-tiket-be/pkg/util"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type OrderHandler interface {
	RegisterRouter(g *echo.Group)
}

type orderHandler struct {
	log          util.LogUtil
	orderService service.OrderService
}

func MakeOrderHandler(log util.LogUtil, orderService service.OrderService) OrderHandler {
	return orderHandler{
		log:          log,
		orderService: orderService,
	}
}

func (h orderHandler) RegisterRouter(g *echo.Group) {
	public := g.Group("/v1")
	public.POST("/webhook/payment/:gateway", h.handleWebhook)
	public.GET("/orders/:order_number/status", h.getOrderStatus)
}

func (h orderHandler) handleWebhook(c echo.Context) error {
	ctx := c.Request().Context()
	gatewayParam := strings.ToLower(c.Param("gateway"))
	var gateway payment.GatewayType

	if gatewayParam == "midtrans" {
		gateway = payment.GatewayMidtrans
	} else if gatewayParam == "xendit" {
		gateway = payment.GatewayXendit
	} else {
		return echo.NewHTTPError(http.StatusBadRequest, "unsupported payment gateway")
	}

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		h.log.Error(ctx, "failed to read webhook body", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to read webhook body")
	}
	defer c.Request().Body.Close()

	if err := h.orderService.HandleWebhook(ctx, gateway, body); err != nil {
		h.log.Warn(ctx, "Webhook processed with warning", zap.Error(err), zap.String("gateway", gatewayParam))
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Webhook received",
	})
}

func (h orderHandler) getOrderStatus(c echo.Context) error {
	orderNumber := c.Param("order_number")

	if orderNumber == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "order_number is required")
	}

	data, err := h.orderService.GetOrderStatus(c.Request().Context(), orderNumber)
	if err != nil {
		if err.Error() == "order not found" {
			return c.JSON(http.StatusNotFound, map[string]interface{}{
				"success": false,
				"message": "Order not found",
			})
		}

		h.log.Error(c.Request().Context(), "getOrderStatus error", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    data,
	})
}
