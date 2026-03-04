package handler

import (
	"rakit-tiket-be/internal/app/app_event/service"
	"rakit-tiket-be/internal/pkg/middleware"

	"github.com/labstack/echo/v4"
)

type HttpHandler interface {
	RegisterRoute(g *echo.Group)
}

type httpHandler struct {
	eventService service.EventService
	eventHandler EventHandler
}

func MakeHttpAdapter(eventService service.EventService, authMiddleware middleware.AuthMiddleware) HttpHandler {
	return httpHandler{
		eventService: eventService,
		eventHandler: MakeEventHandler(eventService, authMiddleware),
	}
}

func (h httpHandler) RegisterRoute(g *echo.Group) {
	h.eventHandler.RegisterRouter(g)
}
