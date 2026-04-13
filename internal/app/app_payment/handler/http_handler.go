package handler

import (
	fileSvc "rakit-tiket-be/internal/app/app_file/service"
	paymentSvc "rakit-tiket-be/internal/app/app_payment/service"
	"rakit-tiket-be/internal/pkg/middleware"
	"rakit-tiket-be/pkg/util"

	"github.com/labstack/echo/v4"
)

type HttpHandler interface {
	RegisterRoute(g *echo.Group)
}

type httpHandler struct {
	bankAccountService    paymentSvc.BankAccountService
	manualTransferService paymentSvc.ManualTransferService
	checkoutService       paymentSvc.CheckoutService
	paymentConfigService  paymentSvc.PaymentConfigService
	fileService           fileSvc.FileService
	authMiddleware        middleware.AuthMiddleware
	log                   util.LogUtil
}

func MakeHttpAdapter(
	log util.LogUtil,
	bankAccountService paymentSvc.BankAccountService,
	manualTransferService paymentSvc.ManualTransferService,
	checkoutService paymentSvc.CheckoutService,
	paymentConfigService paymentSvc.PaymentConfigService,
	fileService fileSvc.FileService,
	authMiddleware middleware.AuthMiddleware,
) HttpHandler {
	return httpHandler{
		log:                   log,
		bankAccountService:    bankAccountService,
		manualTransferService: manualTransferService,
		checkoutService:       checkoutService,
		paymentConfigService:  paymentConfigService,
		fileService:           fileService,
		authMiddleware:        authMiddleware,
	}
}

func (h httpHandler) RegisterRoute(g *echo.Group) {
	paymentHandler := MakePaymentHandler(
		h.log,
		h.bankAccountService,
		h.manualTransferService,
		h.checkoutService,
		h.paymentConfigService,
		h.fileService,
		h.authMiddleware,
	)
	paymentHandler.RegisterRouter(g)
}
