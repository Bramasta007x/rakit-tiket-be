package handler

import (
	"net/http"

	"rakit-tiket-be/internal/app/app_registrant/service"
	model "rakit-tiket-be/pkg/model/app_registrant"

	"github.com/labstack/echo/v4"
)

type RegistrantHandler interface {
	RegisterRouter(g *echo.Group)
}

type registrantHandler struct {
	registrantService service.RegistrantService
}

func MakeRegistrantHandler(registrantService service.RegistrantService) RegistrantHandler {
	return registrantHandler{
		registrantService: registrantService,
	}
}

func (h registrantHandler) RegisterRouter(g *echo.Group) {
	public := g.Group("/v1")

	public.POST("/register", h.register)
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
