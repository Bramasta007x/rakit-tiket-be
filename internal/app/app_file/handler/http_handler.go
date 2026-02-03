package handler

import (
	"rakit-tiket-be/internal/app/app_file/service"
	"rakit-tiket-be/pkg/util"

	"github.com/labstack/echo/v4"
)

type HttpHandler interface {
	RegisterRoute(g *echo.Group)
}

type httpHandler struct {
	log         util.LogUtil
	fileService service.FileService

	fileHandler FileAdapter
}

func MakeHttpAdapter(fileService service.FileService, log util.LogUtil) HttpHandler {
	return httpHandler{
		fileService: fileService,
		fileHandler: MakeFileAdapter(log, fileService),
	}
}

func (h httpHandler) RegisterRoute(g *echo.Group) {
	h.fileHandler.RegisterRouter(g)
}
