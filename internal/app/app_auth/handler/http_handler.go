package handler

import (
	"rakit-tiket-be/internal/app/app_auth/service"
	"rakit-tiket-be/pkg/util"

	"github.com/labstack/echo/v4"
)

type HttpHandler interface {
	RegisterRoute(g *echo.Group)
}

type httpHandler struct {
	authHandler AuthHandler
}

func MakeHttpAdapter(log util.LogUtil, authService service.AuthService) HttpHandler {
	return httpHandler{
		authHandler: MakeAuthHandler(log, authService),
	}
}

func (h httpHandler) RegisterRoute(g *echo.Group) {
	h.authHandler.RegisterRoute(g)
}
