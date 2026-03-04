package handler

import (
	"net/http"

	"rakit-tiket-be/internal/app/app_event/service"
	"rakit-tiket-be/internal/pkg/middleware"
	pubEntity "rakit-tiket-be/pkg/entity"
	entity "rakit-tiket-be/pkg/entity/app_event"

	"github.com/labstack/echo/v4"
)

type EventHandler interface {
	RegisterRouter(g *echo.Group)
}

type eventHandler struct {
	eventService service.EventService
	middleware   middleware.AuthMiddleware
}

func MakeEventHandler(
	eventService service.EventService,
	middleware middleware.AuthMiddleware,
) EventHandler {
	return eventHandler{
		eventService: eventService,
		middleware:   middleware,
	}
}

func (h eventHandler) RegisterRouter(g *echo.Group) {
	restricted := g.Group("/v1/admin")
	restrictedPublic := g.Group("/v1")

	// Public Routes (Biasanya untuk landing page/list event)
	restrictedPublic.GET("/events", h.searchEvents)

	// Admin Routes (Membutuhkan Auth)
	restricted.Use(h.middleware.VerifyToken)
	restricted.Use(h.middleware.RequireAdmin)

	restricted.POST("/events", h.insertEvents)
	restricted.PUT("/events", h.updateEvents)
	restricted.PUT("/event/:id", h.updateEvent)
	restricted.DELETE("/event/:id", h.softDeleteEvent)
}

func (h eventHandler) searchEvents(c echo.Context) error {
	var query entity.EventQuery

	if err := c.Bind(&query); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	events, err := h.eventService.Search(
		c.Request().Context(),
		query,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, events)
}

func (h eventHandler) insertEvents(c echo.Context) error {
	var events entity.Events

	if err := c.Bind(&events); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.eventService.Insert(
		c.Request().Context(),
		events,
	); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, events)
}

func (h eventHandler) updateEvents(c echo.Context) error {
	var events entity.Events

	if err := c.Bind(&events); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.eventService.Update(
		c.Request().Context(),
		events,
	); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, events)
}

func (h eventHandler) updateEvent(c echo.Context) error {
	var event entity.Event

	if err := c.Bind(&event); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Force ID dari parameter URL
	event.ID = pubEntity.UUID(c.Param("id"))

	if err := h.eventService.Update(
		c.Request().Context(),
		entity.Events{event},
	); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, event)
}

func (h eventHandler) softDeleteEvent(c echo.Context) error {
	id := pubEntity.UUID(c.Param("id"))

	if err := h.eventService.SoftDelete(
		c.Request().Context(),
		id,
	); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(
		http.StatusOK,
		map[string]string{
			"message": "Event deleted successfully",
		},
	)
}
