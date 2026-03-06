package handler

import (
	"rakit-tiket-be/internal/app/app_order/service"
	"rakit-tiket-be/pkg/util"

	"github.com/labstack/echo/v4"
)

type HttpHandler interface {
	RegisterRoute(g *echo.Group)
}

type httpHandler struct {
	orderService service.OrderService
	orderHandler OrderHandler
}

func MakeHttpAdapter(log util.LogUtil, orderService service.OrderService) HttpHandler {
	return httpHandler{
		orderService: orderService,
		orderHandler: MakeOrderHandler(log, orderService),
	}
}

func (h httpHandler) RegisterRoute(g *echo.Group) {
	h.orderHandler.RegisterRouter(g)
}
