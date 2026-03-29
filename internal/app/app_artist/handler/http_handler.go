package handler

import (
	"rakit-tiket-be/internal/app/app_artist/service"
	fileService "rakit-tiket-be/internal/app/app_file/service"
	"rakit-tiket-be/internal/pkg/middleware"

	"github.com/labstack/echo/v4"
)

type HttpHandler interface {
	RegisterRoute(g *echo.Group)
}

type httpHandler struct {
	artistService service.ArtistService
	fileService   fileService.FileService
	artistHandler ArtistHandler
}

func MakeHttpAdapter(artistService service.ArtistService, fileService fileService.FileService, authMiddleware middleware.AuthMiddleware) HttpHandler {
	return httpHandler{
		artistService: artistService,
		fileService:   fileService,
		artistHandler: MakeArtistHandler(artistService, fileService, authMiddleware),
	}
}

func (h httpHandler) RegisterRoute(g *echo.Group) {
	h.artistHandler.RegisterRouter(g)
}
