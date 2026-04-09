package handler

import (
	"net/http"
	"time"

	"rakit-tiket-be/internal/app/app_payment/service"
	"rakit-tiket-be/internal/pkg/middleware"
	pubEntity "rakit-tiket-be/pkg/entity"
	"rakit-tiket-be/pkg/util"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type PaymentHandler interface {
	RegisterRouter(g *echo.Group)
}

type paymentHandler struct {
	log                   util.LogUtil
	bankAccountService    service.BankAccountService
	manualTransferService service.ManualTransferService
	authMiddleware        middleware.AuthMiddleware
}

func MakePaymentHandler(
	log util.LogUtil,
	bankAccountService service.BankAccountService,
	manualTransferService service.ManualTransferService,
	authMiddleware middleware.AuthMiddleware,
) PaymentHandler {
	return &paymentHandler{
		log:                   log,
		bankAccountService:    bankAccountService,
		manualTransferService: manualTransferService,
		authMiddleware:        authMiddleware,
	}
}

func (h *paymentHandler) RegisterRouter(g *echo.Group) {
	public := g.Group("/v1")
	public.GET("/bank-accounts", h.getBankAccounts)
	public.POST("/transfers/proof", h.submitTransferProof)

	admin := g.Group("/v1/admin")
	admin.Use(h.authMiddleware.VerifyToken)
	admin.Use(h.authMiddleware.RequireAdmin)
	admin.GET("/transfers/pending", h.getPendingTransfers)
	admin.POST("/transfers/:transfer_id/approve", h.approveTransfer)
	admin.POST("/transfers/:transfer_id/reject", h.rejectTransfer)
	admin.POST("/bank-accounts", h.createBankAccount)
	admin.PUT("/bank-accounts/:bank_account_id", h.updateBankAccount)
	admin.DELETE("/bank-accounts/:bank_account_id", h.deleteBankAccount)
}

func (h *paymentHandler) getBankAccounts(c echo.Context) error {
	ctx := c.Request().Context()

	accounts, err := h.bankAccountService.GetActiveBankAccounts(ctx)
	if err != nil {
		h.log.Error(ctx, "getBankAccounts error", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch bank accounts")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    accounts,
	})
}

func (h *paymentHandler) submitTransferProof(c echo.Context) error {
	ctx := c.Request().Context()

	var req struct {
		OrderID               string  `json:"order_id" validate:"required"`
		BankAccountID         string  `json:"bank_account_id" validate:"required"`
		TransferProofURL      string  `json:"transfer_proof_url" validate:"required"`
		TransferProofFilename *string `json:"transfer_proof_filename"`
		SenderName            string  `json:"sender_name" validate:"required"`
		SenderAccountNumber   *string `json:"sender_account_number"`
		TransferDate          string  `json:"transfer_date" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	transferDate, err := time.Parse(time.RFC3339, req.TransferDate)
	if err != nil {
		transferDate = time.Now()
	}

	serviceReq := service.SubmitTransferProofRequest{
		OrderID:               pubEntity.UUID(req.OrderID),
		BankAccountID:         pubEntity.UUID(req.BankAccountID),
		TransferProofURL:      req.TransferProofURL,
		TransferProofFilename: req.TransferProofFilename,
		SenderName:            req.SenderName,
		SenderAccountNumber:   req.SenderAccountNumber,
		TransferDate:          transferDate,
	}

	transfer, err := h.manualTransferService.SubmitTransferProof(ctx, serviceReq)
	if err != nil {
		h.log.Error(ctx, "submitTransferProof error", zap.Error(err))
		if err == service.ErrTransferAlreadyExists {
			return echo.NewHTTPError(http.StatusConflict, "Transfer proof already submitted for this order")
		}
		if err == service.ErrOrderNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Order not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to submit transfer proof")
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"success": true,
		"message": "Transfer proof submitted successfully",
		"data":    transfer,
	})
}

func (h *paymentHandler) getPendingTransfers(c echo.Context) error {
	ctx := c.Request().Context()

	transfers, err := h.manualTransferService.GetPendingTransfers(ctx)
	if err != nil {
		h.log.Error(ctx, "getPendingTransfers error", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch pending transfers")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    transfers,
		"count":   len(transfers),
	})
}

func (h *paymentHandler) approveTransfer(c echo.Context) error {
	ctx := c.Request().Context()
	transferID := c.Param("transfer_id")

	adminID := "admin"
	if userClaims, ok := c.Get("user").(map[string]interface{}); ok {
		if id, ok := userClaims["id"].(string); ok {
			adminID = id
		}
	}

	var req struct {
		Notes string `json:"notes"`
	}
	c.Bind(&req)

	err := h.manualTransferService.ApproveTransfer(ctx, transferID, adminID, req.Notes)
	if err != nil {
		h.log.Error(ctx, "approveTransfer error", zap.Error(err))
		if err == service.ErrManualTransferNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Transfer not found")
		}
		if err == service.ErrInvalidStatus {
			return echo.NewHTTPError(http.StatusBadRequest, "Transfer is not in pending status")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to approve transfer")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Transfer approved successfully",
	})
}

func (h *paymentHandler) rejectTransfer(c echo.Context) error {
	ctx := c.Request().Context()
	transferID := c.Param("transfer_id")

	adminID := "admin"
	if userClaims, ok := c.Get("user").(map[string]interface{}); ok {
		if id, ok := userClaims["id"].(string); ok {
			adminID = id
		}
	}

	var req struct {
		Notes string `json:"notes"`
	}
	c.Bind(&req)

	err := h.manualTransferService.RejectTransfer(ctx, transferID, adminID, req.Notes)
	if err != nil {
		h.log.Error(ctx, "rejectTransfer error", zap.Error(err))
		if err == service.ErrManualTransferNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Transfer not found")
		}
		if err == service.ErrInvalidStatus {
			return echo.NewHTTPError(http.StatusBadRequest, "Transfer is not in pending status")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to reject transfer")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Transfer rejected",
	})
}

func (h *paymentHandler) createBankAccount(c echo.Context) error {
	ctx := c.Request().Context()

	var req service.CreateBankAccountRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	account, err := h.bankAccountService.CreateBankAccount(ctx, req)
	if err != nil {
		h.log.Error(ctx, "createBankAccount error", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create bank account")
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"success": true,
		"message": "Bank account created successfully",
		"data":    account,
	})
}

func (h *paymentHandler) updateBankAccount(c echo.Context) error {
	ctx := c.Request().Context()
	bankAccountID := c.Param("bank_account_id")

	var req service.UpdateBankAccountRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}
	req.ID = bankAccountID

	err := h.bankAccountService.UpdateBankAccount(ctx, req)
	if err != nil {
		h.log.Error(ctx, "updateBankAccount error", zap.Error(err))
		if err == service.ErrBankAccountNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Bank account not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update bank account")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Bank account updated successfully",
	})
}

func (h *paymentHandler) deleteBankAccount(c echo.Context) error {
	ctx := c.Request().Context()
	bankAccountID := c.Param("bank_account_id")

	err := h.bankAccountService.DeleteBankAccount(ctx, bankAccountID)
	if err != nil {
		h.log.Error(ctx, "deleteBankAccount error", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete bank account")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Bank account deleted successfully",
	})
}
