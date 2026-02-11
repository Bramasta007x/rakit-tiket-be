package handler

import (
	"rakit-tiket-be/internal/app/app_ticket/service"
	"rakit-tiket-be/internal/pkg/middleware"

	"github.com/labstack/echo/v4"
)

type HttpHandler interface {
	RegisterRoute(g *echo.Group)
}

type httpHandler struct {
	ticketService service.TicketService

	ticketHandler TicketHandler
}

func MakeHttpAdapter(
	ticketService service.TicketService,
	authMiddleware middleware.AuthMiddleware,
) HttpHandler {
	return httpHandler{
		ticketService: ticketService,
		ticketHandler: MakeTicketHandler(ticketService, authMiddleware),
	}
}

func (h httpHandler) RegisterRoute(g *echo.Group) {
	h.ticketHandler.RegisterRouter(g)
}
