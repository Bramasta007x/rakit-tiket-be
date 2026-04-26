package handler

import (
	"rakit-tiket-be/internal/app/app_hype/service"
	"rakit-tiket-be/internal/pkg/middleware"
	"rakit-tiket-be/pkg/util"

	"github.com/labstack/echo/v4"
)

type HttpHandler interface {
	RegisterRoute(g *echo.Group)
}

type httpHandler struct {
	hypeService service.HypeService
	hypeHandler HypeHandler
}

func MakeHttpAdapter(log util.LogUtil, hypeService service.HypeService, authMiddleware middleware.AuthMiddleware) HttpHandler {
	return httpHandler{
		hypeService: hypeService,
		hypeHandler: MakeHypeHandler(log, hypeService, authMiddleware),
	}
}

func (h httpHandler) RegisterRoute(g *echo.Group) {
	h.hypeHandler.RegisterRouter(g)
}
