package handler

import (
	"rakit-tiket-be/internal/app/app_landing_page/service"

	"github.com/labstack/echo/v4"
)

type HttpHandler interface {
	RegisterRoute(g *echo.Group)
}

type httpHandler struct {
	landingPageService service.LandingPageService

	landingPageHandler LandingPageHandler
}

func MakeHttpAdapter(
	landingPageService service.LandingPageService,
) HttpHandler {
	return httpHandler{
		landingPageService: landingPageService,
		landingPageHandler: MakeLandingPageHandler(landingPageService),
	}
}

func (h httpHandler) RegisterRoute(g *echo.Group) {
	h.landingPageHandler.RegisterRouter(g)
}
