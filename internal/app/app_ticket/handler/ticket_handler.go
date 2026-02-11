package handler

import (
	"net/http"

	"rakit-tiket-be/internal/app/app_ticket/service"
	"rakit-tiket-be/internal/pkg/middleware"
	pubEntity "rakit-tiket-be/pkg/entity"
	entity "rakit-tiket-be/pkg/entity/app_ticket"

	"github.com/labstack/echo/v4"
)

type TicketHandler interface {
	RegisterRouter(g *echo.Group)
}

type ticketHandler struct {
	ticketService service.TicketService
	middleware    middleware.AuthMiddleware
}

func MakeTicketHandler(
	ticketService service.TicketService,
	middleware middleware.AuthMiddleware,
) ticketHandler {
	return ticketHandler{
		ticketService: ticketService,
		middleware:    middleware,
	}
}

func (h ticketHandler) RegisterRouter(g *echo.Group) {
	restricted := g.Group("/v1/admin")
	restrictedPublic := g.Group("/v1")

	restrictedPublic.GET("/tickets", h.searchTickets)

	restricted.Use(h.middleware.VerifyToken)
	restricted.Use(h.middleware.RequireAdmin)

	restricted.POST("/tickets", h.insertTickets)
	restricted.PUT("/tickets", h.updateTickets)
	restricted.PUT("/ticket/:id", h.updateTicket)

	restricted.DELETE("/ticket/:id", h.softDeleteTicket)
}

func (h ticketHandler) searchTickets(c echo.Context) error {
	var query entity.TicketQuery

	if err := c.Bind(&query); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	tickets, err := h.ticketService.Search(
		c.Request().Context(),
		query,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, tickets)
}

func (h ticketHandler) insertTickets(c echo.Context) error {
	var tickets entity.Tickets

	if err := c.Bind(&tickets); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.ticketService.Insert(
		c.Request().Context(),
		tickets,
	); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, tickets)
}

func (h ticketHandler) updateTickets(c echo.Context) error {
	var tickets entity.Tickets

	if err := c.Bind(&tickets); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.ticketService.Update(
		c.Request().Context(),
		tickets,
	); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, tickets)
}

func (h ticketHandler) updateTicket(c echo.Context) error {
	var ticket entity.Ticket

	if err := c.Bind(&ticket); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Force ID dari URL
	ticket.ID = pubEntity.UUID(c.Param("id"))

	if err := h.ticketService.Update(
		c.Request().Context(),
		entity.Tickets{ticket},
	); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, ticket)
}

func (h ticketHandler) softDeleteTicket(c echo.Context) error {
	id := pubEntity.UUID(c.Param("id"))

	if err := h.ticketService.SoftDelete(
		c.Request().Context(),
		id,
	); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(
		http.StatusOK,
		map[string]string{
			"message": "Ticket deleted successfully",
		},
	)
}
