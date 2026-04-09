package handler

import (
	"rakit-tiket-be/internal/app/app_payment/service"
	"rakit-tiket-be/internal/pkg/middleware"
	"rakit-tiket-be/pkg/util"

	"github.com/labstack/echo/v4"
)

type HttpHandler interface {
	RegisterRoute(g *echo.Group)
}

type httpHandler struct {
	bankAccountService    service.BankAccountService
	manualTransferService service.ManualTransferService
	authMiddleware        middleware.AuthMiddleware
	log                   util.LogUtil
}

func MakeHttpAdapter(
	log util.LogUtil,
	bankAccountService service.BankAccountService,
	manualTransferService service.ManualTransferService,
	authMiddleware middleware.AuthMiddleware,
) HttpHandler {
	return httpHandler{
		log:                   log,
		bankAccountService:    bankAccountService,
		manualTransferService: manualTransferService,
		authMiddleware:        authMiddleware,
	}
}

func (h httpHandler) RegisterRoute(g *echo.Group) {
	paymentHandler := MakePaymentHandler(
		h.log,
		h.bankAccountService,
		h.manualTransferService,
		h.authMiddleware,
	)
	paymentHandler.RegisterRouter(g)
}
