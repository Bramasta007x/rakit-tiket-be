package handler

import (
	"rakit-tiket-be/internal/app/app_registrant/service"
	"rakit-tiket-be/internal/pkg/middleware"

	"github.com/labstack/echo/v4"
)

type HttpHandler interface {
	RegisterRoute(g *echo.Group)
}

type httpHandler struct {
	registrantService service.RegistrantService
	registrantHandler RegistrantHandler
}

func MakeHttpAdapter(
	registrantService service.RegistrantService,
	middleware middleware.AuthMiddleware,
) HttpHandler {
	return httpHandler{
		registrantService: registrantService,
		registrantHandler: MakeRegistrantHandler(registrantService, middleware),
	}
}

func (h httpHandler) RegisterRoute(g *echo.Group) {
	h.registrantHandler.RegisterRouter(g)
}
