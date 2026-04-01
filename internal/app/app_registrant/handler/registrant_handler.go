package handler

import (
	"net/http"
	"os"
	"path/filepath"

	"rakit-tiket-be/internal/app/app_registrant/service"
	"rakit-tiket-be/internal/pkg/middleware"
	model "rakit-tiket-be/pkg/model/app_registrant"

	"github.com/labstack/echo/v4"
	"gitlab.com/threetopia/envgo"
)

type RegistrantHandler interface {
	RegisterRouter(g *echo.Group)
}

type registrantHandler struct {
	registrantService service.RegistrantService
	middleware        middleware.AuthMiddleware
}

func MakeRegistrantHandler(
	registrantService service.RegistrantService,
	middleware middleware.AuthMiddleware,
) RegistrantHandler {
	return registrantHandler{
		registrantService: registrantService,
		middleware:        middleware,
	}
}

func (h registrantHandler) RegisterRouter(g *echo.Group) {
	restricted := g.Group("/v1/admin")
	restrictedPublic := g.Group("/v1")

	restrictedPublic.POST("/register", h.register)

	restricted.Use(h.middleware.VerifyToken)
	restricted.Use(h.middleware.RequireAdmin)

	restricted.GET("/registrants", h.list)
	restricted.GET("/ticket/:filename", h.downloadTicket)
}

func (h registrantHandler) register(c echo.Context) error {
	var req model.RegisterRequest

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	resp, err := h.registrantService.Register(c.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, resp)
}

func (h registrantHandler) list(c echo.Context) error {
	var req model.SearchRegistrantsRequestModel

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	httpCode, resp := h.registrantService.List(c.Request().Context(), req)

	return c.JSON(httpCode, resp)
}

func (h registrantHandler) downloadTicket(c echo.Context) error {
	filename := c.Param("filename")
	if filename == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "filename is required")
	}

	ticketDir := filepath.Join(envgo.GetString("APP_FILE_PATH", "./assets/app_file"), "tickets")
	filePath := filepath.Join(ticketDir, filename)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return echo.NewHTTPError(http.StatusNotFound, "ticket file not found")
	}

	return c.File(filePath)
}
