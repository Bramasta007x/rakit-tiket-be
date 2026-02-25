package handler

import (
	"rakit-tiket-be/internal/app/app_order/service"

	"github.com/labstack/echo/v4"
)

type HttpHandler interface {
	RegisterRoute(g *echo.Group)
}

type httpHandler struct {
	orderService service.OrderService
	orderHandler OrderHandler
}

func MakeHttpAdapter(orderService service.OrderService) HttpHandler {
	return httpHandler{
		orderService: orderService,
		orderHandler: MakeOrderHandler(orderService),
	}
}

func (h httpHandler) RegisterRoute(g *echo.Group) {
	h.orderHandler.RegisterRouter(g)
}
