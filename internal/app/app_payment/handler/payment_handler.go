package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	fileSvc "rakit-tiket-be/internal/app/app_file/service"
	paymentSvc "rakit-tiket-be/internal/app/app_payment/service"
	"rakit-tiket-be/internal/pkg/middleware"
	pubEntity "rakit-tiket-be/pkg/entity"
	fileEntity "rakit-tiket-be/pkg/entity/app_file"
	"rakit-tiket-be/pkg/util"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type PaymentHandler interface {
	RegisterRouter(g *echo.Group)
}

type paymentHandler struct {
	log                   util.LogUtil
	bankAccountService    paymentSvc.BankAccountService
	manualTransferService paymentSvc.ManualTransferService
	checkoutService       paymentSvc.CheckoutService
	paymentConfigService  paymentSvc.PaymentConfigService
	fileService           fileSvc.FileService
	authMiddleware        middleware.AuthMiddleware
}

func MakePaymentHandler(
	log util.LogUtil,
	bankAccountService paymentSvc.BankAccountService,
	manualTransferService paymentSvc.ManualTransferService,
	checkoutService paymentSvc.CheckoutService,
	paymentConfigService paymentSvc.PaymentConfigService,
	fileService fileSvc.FileService,
	authMiddleware middleware.AuthMiddleware,
) PaymentHandler {
	return &paymentHandler{
		log:                   log,
		bankAccountService:    bankAccountService,
		manualTransferService: manualTransferService,
		checkoutService:       checkoutService,
		paymentConfigService:  paymentConfigService,
		fileService:           fileService,
		authMiddleware:        authMiddleware,
	}
}

func (h *paymentHandler) RegisterRouter(g *echo.Group) {
	public := g.Group("/v1")

	public.GET("/bank-accounts", h.getBankAccounts)
	public.POST("/transfers/proof", h.submitTransferProof)
	public.POST("/checkout/:order_id", h.initiateCheckout)
	public.GET("/payment-options", h.getPaymentOptions)

	admin := g.Group("/v1/admin")
	admin.Use(h.authMiddleware.VerifyToken)
	admin.Use(h.authMiddleware.RequireAdmin)

	admin.GET("/transfers/pending", h.getPendingTransfers)
	admin.POST("/transfers/:transfer_id/approve", h.approveTransfer)
	admin.POST("/transfers/:transfer_id/reject", h.rejectTransfer)

	admin.POST("/bank-accounts", h.createBankAccount)
	admin.PUT("/bank-accounts/:bank_account_id", h.updateBankAccount)
	admin.DELETE("/bank-accounts/:bank_account_id", h.deleteBankAccount)

	admin.GET("/gateways", h.getAllGateways)
	admin.POST("/gateways/:code/activate", h.activateGateway)
	admin.POST("/gateways/:code/deactivate", h.deactivateGateway)
	admin.PUT("/gateways/:code/display-order", h.setGatewayDisplayOrder)

	admin.POST("/manual-transfer/enable", h.enableManualTransfer)
	admin.POST("/manual-transfer/disable", h.disableManualTransfer)
	admin.PUT("/manual-transfer/display-order", h.setManualTransferDisplayOrder)
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
	contentType := c.Request().Header.Get("Content-Type")

	if strings.Contains(contentType, "multipart/form-data") {
		return h.handleMultipartTransferProof(c)
	}

	return h.handleJSONTransferProof(c)
}

func (h *paymentHandler) handleMultipartTransferProof(c echo.Context) error {
	ctx := c.Request().Context()

	jsonData := c.FormValue("data")
	if jsonData == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing 'data' field in form")
	}

	var req struct {
		OrderID             string  `json:"order_id"`
		BankAccountID       string  `json:"bank_account_id"`
		SenderName          string  `json:"sender_name"`
		SenderAccountNumber *string `json:"sender_account_number"`
		TransferDate        string  `json:"transfer_date"`
	}

	if err := json.Unmarshal([]byte(jsonData), &req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid JSON format in 'data' field: "+err.Error())
	}

	if req.OrderID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "order_id is required")
	}
	if req.BankAccountID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "bank_account_id is required")
	}
	if req.SenderName == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "sender_name is required")
	}

	transferProofURL := ""
	transferProofFilename := ""

	fileHeader, err := c.FormFile("transfer_proof")
	if err == nil {
		src, err := fileHeader.Open()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to open uploaded file: "+err.Error())
		}
		defer src.Close()

		fileBytes, err := io.ReadAll(src)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to read uploaded file: "+err.Error())
		}

		relationID := pubEntity.MakeUUID(req.OrderID, time.Now().String())
		fileEntity := &fileEntity.FileEntity{
			Name:        fileHeader.Filename,
			Description: fmt.Sprintf("Transfer proof for order %s", req.OrderID),
			Data:        fileBytes,
			RelationEntity: pubEntity.MakeRelationEntity(
				relationID,
				"transfer_proof",
			),
		}

		if err := h.fileService.Insert(ctx, fileEntity); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to save transfer proof file: "+err.Error())
		}

		filePathFolder := h.fileService.GetFilePath()
		transferProofURL = fmt.Sprintf("%s/%s/%s/%s.ref", filePathFolder, fileEntity.RelationSource, fileEntity.RelationID.String(), fileEntity.ID.String())
		transferProofFilename = fileHeader.Filename

	} else if err != http.ErrMissingFile {
		return echo.NewHTTPError(http.StatusBadRequest, "Error reading transfer_proof file: "+err.Error())
	}

	if transferProofURL == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "transfer_proof file is required")
	}

	transferDate, err := time.Parse(time.RFC3339, req.TransferDate)
	if err != nil {
		transferDate = time.Now()
	}

	serviceReq := paymentSvc.SubmitTransferProofRequest{
		OrderID:               pubEntity.UUID(req.OrderID),
		BankAccountID:         pubEntity.UUID(req.BankAccountID),
		TransferProofURL:      transferProofURL,
		TransferProofFilename: &transferProofFilename,
		SenderName:            req.SenderName,
		SenderAccountNumber:   req.SenderAccountNumber,
		TransferDate:          transferDate,
	}

	transfer, err := h.manualTransferService.SubmitTransferProof(ctx, serviceReq)
	if err != nil {
		h.log.Error(ctx, "submitTransferProof error", zap.Error(err))
		if err == paymentSvc.ErrTransferAlreadyExists {
			return echo.NewHTTPError(http.StatusConflict, "Transfer proof already submitted for this order")
		}
		if err == paymentSvc.ErrOrderNotFound {
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

func (h *paymentHandler) handleJSONTransferProof(c echo.Context) error {
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

	serviceReq := paymentSvc.SubmitTransferProofRequest{
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
		if err == paymentSvc.ErrTransferAlreadyExists {
			return echo.NewHTTPError(http.StatusConflict, "Transfer proof already submitted for this order")
		}
		if err == paymentSvc.ErrOrderNotFound {
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

	adminID := ""
	if userID, ok := c.Get("user_id").(string); ok {
		adminID = userID
	}

	var req struct {
		Notes string `json:"notes"`
	}
	c.Bind(&req)

	err := h.manualTransferService.ApproveTransfer(ctx, transferID, adminID, req.Notes)
	if err != nil {
		h.log.Error(ctx, "approveTransfer error", zap.Error(err))
		if err == paymentSvc.ErrManualTransferNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Transfer not found")
		}
		if err == paymentSvc.ErrInvalidStatus {
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

	adminID := ""
	if userID, ok := c.Get("user_id").(string); ok {
		adminID = userID
	}

	var req struct {
		Notes string `json:"notes"`
	}
	c.Bind(&req)

	err := h.manualTransferService.RejectTransfer(ctx, transferID, adminID, req.Notes)
	if err != nil {
		h.log.Error(ctx, "rejectTransfer error", zap.Error(err))
		if err == paymentSvc.ErrManualTransferNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Transfer not found")
		}
		if err == paymentSvc.ErrInvalidStatus {
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

	var req paymentSvc.CreateBankAccountRequest
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

	var req paymentSvc.UpdateBankAccountRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}
	req.ID = bankAccountID

	err := h.bankAccountService.UpdateBankAccount(ctx, req)
	if err != nil {
		h.log.Error(ctx, "updateBankAccount error", zap.Error(err))
		if err == paymentSvc.ErrBankAccountNotFound {
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

func (h *paymentHandler) initiateCheckout(c echo.Context) error {
	orderID := c.Param("order_id")
	if orderID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "order_id is required")
	}

	var req struct {
		PaymentType string `json:"payment_type"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if req.PaymentType == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "payment_type is required")
	}

	data, err := h.checkoutService.InitiateCheckout(c.Request().Context(), orderID, req.PaymentType)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.JSON(http.StatusNotFound, map[string]interface{}{
				"success": false,
				"message": err.Error(),
			})
		}
		if strings.Contains(err.Error(), "expired") {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": err.Error(),
			})
		}
		h.log.Error(c.Request().Context(), "initiateCheckout error", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

func (h *paymentHandler) getPaymentOptions(c echo.Context) error {
	ctx := c.Request().Context()

	options, err := h.checkoutService.GetActivePaymentOptions(ctx)
	if err != nil {
		h.log.Error(ctx, "getPaymentOptions error", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch payment options")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    options,
	})
}

func (h *paymentHandler) getAllGateways(c echo.Context) error {
	ctx := c.Request().Context()

	gateways, err := h.paymentConfigService.GetAllGateways(ctx)
	if err != nil {
		h.log.Error(ctx, "getAllGateways error", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch gateways")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    gateways,
	})
}

func (h *paymentHandler) activateGateway(c echo.Context) error {
	ctx := c.Request().Context()
	code := strings.ToUpper(c.Param("code"))

	err := h.paymentConfigService.ActivateGateway(ctx, code)
	if err != nil {
		if err == paymentSvc.ErrGatewayNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Gateway not found")
		}
		h.log.Error(ctx, "activateGateway error", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Gateway %s activated successfully", code),
	})
}

func (h *paymentHandler) deactivateGateway(c echo.Context) error {
	ctx := c.Request().Context()
	code := strings.ToUpper(c.Param("code"))

	err := h.paymentConfigService.DeactivateGateway(ctx, code)
	if err != nil {
		if err == paymentSvc.ErrGatewayNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Gateway not found")
		}
		h.log.Error(ctx, "deactivateGateway error", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Gateway %s deactivated successfully", code),
	})
}

func (h *paymentHandler) setGatewayDisplayOrder(c echo.Context) error {
	ctx := c.Request().Context()
	code := strings.ToUpper(c.Param("code"))

	var req struct {
		DisplayOrder int `json:"display_order"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	err := h.paymentConfigService.SetGatewayDisplayOrder(ctx, code, req.DisplayOrder)
	if err != nil {
		h.log.Error(ctx, "setGatewayDisplayOrder error", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update display order")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Display order updated successfully",
	})
}

func (h *paymentHandler) enableManualTransfer(c echo.Context) error {
	ctx := c.Request().Context()

	err := h.paymentConfigService.SetManualTransferEnabled(ctx, true)
	if err != nil {
		h.log.Error(ctx, "enableManualTransfer error", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to enable manual transfer")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Manual transfer enabled successfully",
	})
}

func (h *paymentHandler) disableManualTransfer(c echo.Context) error {
	ctx := c.Request().Context()

	err := h.paymentConfigService.SetManualTransferEnabled(ctx, false)
	if err != nil {
		h.log.Error(ctx, "disableManualTransfer error", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to disable manual transfer")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Manual transfer disabled successfully",
	})
}

func (h *paymentHandler) setManualTransferDisplayOrder(c echo.Context) error {
	ctx := c.Request().Context()

	var req struct {
		DisplayOrder int `json:"display_order"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	err := h.paymentConfigService.SetManualTransferDisplayOrder(ctx, req.DisplayOrder)
	if err != nil {
		h.log.Error(ctx, "setManualTransferDisplayOrder error", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update display order")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Manual transfer display order updated successfully",
	})
}
