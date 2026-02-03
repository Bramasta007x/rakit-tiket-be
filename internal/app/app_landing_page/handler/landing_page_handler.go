package handler

import (
	"net/http"

	"rakit-tiket-be/internal/app/app_landing_page/service"
	pubEntity "rakit-tiket-be/pkg/entity"
	entity "rakit-tiket-be/pkg/entity/app_landing_page"

	"github.com/labstack/echo/v4"
)

type LandingPageHandler interface {
	RegisterRouter(g *echo.Group)
}

type landingPageHandler struct {
	landingPageService service.LandingPageService
}

func MakeLandingPageHandler(landingPageService service.LandingPageService) landingPageHandler {
	return landingPageHandler{
		landingPageService: landingPageService,
	}
}

func (h landingPageHandler) RegisterRouter(g *echo.Group) {
	restricted := g.Group("/v1/admin")

	restricted.GET("/landing-pages", h.searchLandingPages)

	restricted.POST("/landing-pages", h.insertLandingPages)
	restricted.POST("/landing-page", h.insertLandingPage)

	restricted.PUT("/landing-pages", h.updateLandingPages)
	restricted.PUT("/landing-page/:id", h.updateLandingPage)

	restricted.DELETE("/landing-page/:id", h.softDeleteLandingPage)
}

func (h landingPageHandler) searchLandingPages(c echo.Context) error {
	var query entity.LandingPageQuery
	if err := c.Bind(&query); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	pages, err := h.landingPageService.Search(c.Request().Context(), query)
	if err != nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			err.Error(),
		)
	}

	return c.JSON(http.StatusOK, pages)
}

func (h landingPageHandler) insertLandingPages(c echo.Context) error {
	var pages entity.LandingPages
	if err := c.Bind(&pages); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.landingPageService.Insert(
		c.Request().Context(),
		pages,
	); err != nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			err.Error(),
		)
	}

	return c.JSON(http.StatusCreated, pages)
}

func (h landingPageHandler) insertLandingPage(c echo.Context) error {
	var page entity.LandingPage
	if err := c.Bind(&page); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.landingPageService.Insert(
		c.Request().Context(),
		entity.LandingPages{page},
	); err != nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			err.Error(),
		)
	}

	return c.JSON(http.StatusCreated, page)
}

func (h landingPageHandler) updateLandingPages(c echo.Context) error {
	var pages entity.LandingPages
	if err := c.Bind(&pages); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.landingPageService.Update(
		c.Request().Context(),
		pages,
	); err != nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			err.Error(),
		)
	}

	return c.JSON(http.StatusOK, pages)
}

func (h landingPageHandler) updateLandingPage(c echo.Context) error {
	var page entity.LandingPage

	if err := c.Bind(&page); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.landingPageService.Update(
		c.Request().Context(),
		entity.LandingPages{page},
	); err != nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			err.Error(),
		)
	}

	return c.JSON(http.StatusOK, page)
}

func (h landingPageHandler) softDeleteLandingPage(c echo.Context) error {
	id := pubEntity.UUID(c.Param("id"))

	if err := h.landingPageService.SoftDelete(
		c.Request().Context(),
		id,
	); err != nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			err.Error(),
		)
	}

	return c.JSON(
		http.StatusOK,
		map[string]string{"message": "Landing page deleted successfully"},
	)
}
