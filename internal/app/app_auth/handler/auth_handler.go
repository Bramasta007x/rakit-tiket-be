package handler

import (
	"net/http"

	"rakit-tiket-be/internal/app/app_auth/service"
	model "rakit-tiket-be/pkg/model/app_auth"
	"rakit-tiket-be/pkg/util"

	"github.com/labstack/echo/v4"
)

type AuthHandler interface {
	RegisterRoute(g *echo.Group)
}

type authHandler struct {
	log         util.LogUtil
	authService service.AuthService
}

func MakeAuthHandler(log util.LogUtil, authService service.AuthService) AuthHandler {
	return authHandler{
		log:         log,
		authService: authService,
	}
}

func (h authHandler) RegisterRoute(g *echo.Group) {

	restricted := g.Group("/v1/admin")

	restricted.POST("/login", h.login)
}

func (h authHandler) login(c echo.Context) error {
	var req model.LoginRequestModel

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	token, err := h.authService.Login(c.Request().Context(), req.Email, req.Password)

	if err != nil {
		// Jangan expose error detail database/bcrypt ke client untuk keamanan
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid email or password")
	}

	return c.JSON(model.MakeLoginResponseModel(http.StatusOK, token))
}
