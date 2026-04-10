package handler

import (
	"net/http"

	"rakit-tiket-be/internal/app/app_auth/service"
	"rakit-tiket-be/internal/pkg/middleware"
	model "rakit-tiket-be/pkg/model/app_auth"
	"rakit-tiket-be/pkg/util"

	"github.com/labstack/echo/v4"
)

type AuthHandler interface {
	RegisterRoute(g *echo.Group)
}

type authHandler struct {
	log            util.LogUtil
	authService    service.AuthService
	authMiddleware middleware.AuthMiddleware
}

func MakeAuthHandler(log util.LogUtil, authService service.AuthService, authMiddleware middleware.AuthMiddleware) AuthHandler {
	return authHandler{
		log:            log,
		authService:    authService,
		authMiddleware: authMiddleware,
	}
}

func (h authHandler) RegisterRoute(g *echo.Group) {

	restricted := g.Group("/v1/admin")

	restricted.POST("/login", h.login)

	authGroup := g.Group("/v1/admin")
	authGroup.Use(h.authMiddleware.VerifyToken)
	authGroup.GET("/me", h.getCurrentUser)
}

func (h authHandler) login(c echo.Context) error {
	var req model.LoginRequestModel

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	token, err := h.authService.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid email or password")
	}

	return c.JSON(model.MakeLoginResponseModel(http.StatusOK, token))
}

func (h authHandler) getCurrentUser(c echo.Context) error {
	userID, _ := c.Get("user_id").(string)
	role, _ := c.Get("role").(string)

	if userID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token: user_id not found")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"id":   userID,
			"role": role,
		},
	})
}
