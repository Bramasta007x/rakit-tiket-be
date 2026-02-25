package handler

import (
	"rakit-tiket-be/internal/app/app_registrant/service"

	"github.com/labstack/echo/v4"
)

type HttpHandler interface {
	RegisterRoute(g *echo.Group)
}

type httpHandler struct {
	registrantService service.RegistrantService
	registrantHandler RegistrantHandler
}

func MakeHttpAdapter(registrantService service.RegistrantService) HttpHandler {
	return httpHandler{
		registrantService: registrantService,
		registrantHandler: MakeRegistrantHandler(registrantService),
	}
}

func (h httpHandler) RegisterRoute(g *echo.Group) {
	h.registrantHandler.RegisterRouter(g)
}
