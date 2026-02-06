package handler

import (
	fileService "rakit-tiket-be/internal/app/app_file/service"
	"rakit-tiket-be/internal/app/app_landing_page/service"
	"rakit-tiket-be/internal/pkg/middleware"

	"github.com/labstack/echo/v4"
)

type HttpHandler interface {
	RegisterRoute(g *echo.Group)
}

type httpHandler struct {
	landingPageService service.LandingPageService

	landingPageHandler LandingPageHandler
}

func MakeHttpAdapter(landingPageService service.LandingPageService, fileService fileService.FileService, authMiddleware middleware.AuthMiddleware) HttpHandler {
	return httpHandler{
		landingPageService: landingPageService,
		landingPageHandler: MakeLandingPageHandler(landingPageService, fileService, authMiddleware),
	}
}

func (h httpHandler) RegisterRoute(g *echo.Group) {
	h.landingPageHandler.RegisterRouter(g)
}
