package service

import (
	"context"
	"database/sql"
	"errors"

	"rakit-tiket-be/internal/app/app_payment/dao"
	"rakit-tiket-be/pkg/entity/app_payment"
	"rakit-tiket-be/pkg/util"

	"go.uber.org/zap"
)

var (
	ErrGatewayNotFound = errors.New("gateway not found")
)

type PaymentOption struct {
	Type         string `json:"type"`
	Code         string `json:"code"`
	GatewayName  string `json:"gateway_name,omitempty"`
	DisplayOrder int    `json:"display_order"`
}

type PaymentConfigService interface {
	GetActivePaymentOptions(ctx context.Context) ([]PaymentOption, error)
	GetActiveGateway(ctx context.Context) (*app_payment.PaymentGateway, error)
	IsManualTransferEnabled(ctx context.Context) (bool, error)
	GetAllGateways(ctx context.Context) (app_payment.PaymentGateways, error)
	ActivateGateway(ctx context.Context, code string) error
	DeactivateGateway(ctx context.Context, code string) error
	SetManualTransferEnabled(ctx context.Context, enabled bool) error
	SetGatewayDisplayOrder(ctx context.Context, code string, order int) error
	SetManualTransferDisplayOrder(ctx context.Context, order int) error
	GetManualTransferSetting(ctx context.Context) (*ManualTransferSetting, error)
}

type ManualTransferSetting struct {
	Enabled      bool `json:"enabled"`
	DisplayOrder int  `json:"display_order"`
}

type paymentConfigService struct {
	log   util.LogUtil
	sqlDB *sql.DB
}

func MakePaymentConfigService(log util.LogUtil, sqlDB *sql.DB) PaymentConfigService {
	return &paymentConfigService{
		log:   log,
		sqlDB: sqlDB,
	}
}

func (s *paymentConfigService) GetActivePaymentOptions(ctx context.Context) ([]PaymentOption, error) {
	dbTrx := dao.NewTransactionPayment(ctx, s.log, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	var options []PaymentOption

	isActive := true
	gateways, err := dbTrx.GetGatewayDAO().Search(ctx, app_payment.GatewayQuery{
		IsActive:  &isActive,
		IsEnabled: &isActive,
	})
	if err != nil {
		s.log.Error(ctx, "paymentConfigService.GetActiveGateway", zap.Error(err))
		return nil, err
	}

	for _, gateway := range gateways {
		options = append(options, PaymentOption{
			Type:         "GATEWAY",
			Code:         gateway.Code,
			GatewayName:  gateway.Name,
			DisplayOrder: gateway.DisplayOrder,
		})
	}

	manualEnabled, err := s.IsManualTransferEnabled(ctx)
	if err != nil {
		s.log.Error(ctx, "paymentConfigService.IsManualTransferEnabled", zap.Error(err))
		return nil, err
	}

	if manualEnabled {
		manualSetting, _ := s.GetManualTransferSetting(ctx)
		displayOrder := 0
		if manualSetting != nil {
			displayOrder = manualSetting.DisplayOrder
		}
		options = append(options, PaymentOption{
			Type:         "MANUAL",
			Code:         "MANUAL",
			GatewayName:  "Transfer Manual",
			DisplayOrder: displayOrder,
		})
	}

	return options, nil
}

func (s *paymentConfigService) GetActiveGateway(ctx context.Context) (*app_payment.PaymentGateway, error) {
	dbTrx := dao.NewTransactionPayment(ctx, s.log, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	isActive := true
	isEnabled := true
	gateways, err := dbTrx.GetGatewayDAO().Search(ctx, app_payment.GatewayQuery{
		IsActive:  &isActive,
		IsEnabled: &isEnabled,
	})
	if err != nil {
		s.log.Error(ctx, "paymentConfigService.GetActiveGateway", zap.Error(err))
		return nil, err
	}

	if len(gateways) == 0 {
		return nil, nil
	}

	return &gateways[0], nil
}

func (s *paymentConfigService) IsManualTransferEnabled(ctx context.Context) (bool, error) {
	dbTrx := dao.NewTransactionPayment(ctx, s.log, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	setting, err := dbTrx.GetPaymentSettingDAO().GetByKey(ctx, app_payment.SettingKeyManualTransferEnabled)
	if err != nil {
		s.log.Error(ctx, "paymentConfigService.IsManualTransferEnabled", zap.Error(err))
		return false, err
	}

	if setting == nil {
		return false, nil
	}

	return setting.SettingValue, nil
}

func (s *paymentConfigService) GetManualTransferSetting(ctx context.Context) (*ManualTransferSetting, error) {
	dbTrx := dao.NewTransactionPayment(ctx, s.log, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	setting, err := dbTrx.GetPaymentSettingDAO().GetByKey(ctx, app_payment.SettingKeyManualTransferEnabled)
	if err != nil {
		s.log.Error(ctx, "paymentConfigService.GetManualTransferSetting", zap.Error(err))
		return nil, err
	}

	if setting == nil {
		return nil, nil
	}

	return &ManualTransferSetting{
		Enabled:      setting.SettingValue,
		DisplayOrder: setting.DisplayOrder,
	}, nil
}

func (s *paymentConfigService) GetAllGateways(ctx context.Context) (app_payment.PaymentGateways, error) {
	dbTrx := dao.NewTransactionPayment(ctx, s.log, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	gateways, err := dbTrx.GetGatewayDAO().Search(ctx, app_payment.GatewayQuery{})
	if err != nil {
		s.log.Error(ctx, "paymentConfigService.GetAllGateways", zap.Error(err))
		return nil, err
	}

	return gateways, nil
}

func (s *paymentConfigService) ActivateGateway(ctx context.Context, code string) error {
	dbTrx := dao.NewTransactionPayment(ctx, s.log, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	gateways, err := dbTrx.GetGatewayDAO().Search(ctx, app_payment.GatewayQuery{
		Codes: []string{code},
	})
	if err != nil {
		s.log.Error(ctx, "paymentConfigService.ActivateGateway.Search", zap.Error(err))
		return err
	}

	if len(gateways) == 0 {
		return ErrGatewayNotFound
	}

	gateway := gateways[0]

	if !gateway.IsEnabled {
		return errors.New("gateway is not enabled")
	}

	if err := dbTrx.GetGatewayDAO().DeactivateAll(ctx); err != nil {
		s.log.Error(ctx, "paymentConfigService.ActivateGateway.DeactivateAll", zap.Error(err))
		return err
	}

	if err := dbTrx.GetGatewayDAO().SetActiveGateway(ctx, code); err != nil {
		s.log.Error(ctx, "paymentConfigService.ActivateGateway.SetActive", zap.Error(err))
		return err
	}

	return dbTrx.GetSqlTx().Commit()
}

func (s *paymentConfigService) DeactivateGateway(ctx context.Context, code string) error {
	dbTrx := dao.NewTransactionPayment(ctx, s.log, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	gateways, err := dbTrx.GetGatewayDAO().Search(ctx, app_payment.GatewayQuery{
		Codes: []string{code},
	})
	if err != nil {
		s.log.Error(ctx, "paymentConfigService.DeactivateGateway.Search", zap.Error(err))
		return err
	}

	if len(gateways) == 0 {
		return ErrGatewayNotFound
	}

	gateway := gateways[0]
	gateway.IsActive = false

	if err := dbTrx.GetGatewayDAO().Update(ctx, &gateway); err != nil {
		s.log.Error(ctx, "paymentConfigService.DeactivateGateway.Update", zap.Error(err))
		return err
	}

	return dbTrx.GetSqlTx().Commit()
}

func (s *paymentConfigService) SetManualTransferEnabled(ctx context.Context, enabled bool) error {
	dbTrx := dao.NewTransactionPayment(ctx, s.log, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	setting, err := dbTrx.GetPaymentSettingDAO().GetByKey(ctx, app_payment.SettingKeyManualTransferEnabled)
	if err != nil {
		s.log.Error(ctx, "paymentConfigService.SetManualTransferEnabled.GetByKey", zap.Error(err))
		return err
	}

	if setting == nil {
		if err := dbTrx.GetPaymentSettingDAO().Upsert(ctx, app_payment.SettingKeyManualTransferEnabled, enabled, 1); err != nil {
			s.log.Error(ctx, "paymentConfigService.SetManualTransferEnabled.Upsert", zap.Error(err))
			return err
		}
	} else {
		if err := dbTrx.GetPaymentSettingDAO().UpdateSettingValue(ctx, app_payment.SettingKeyManualTransferEnabled, enabled); err != nil {
			s.log.Error(ctx, "paymentConfigService.SetManualTransferEnabled.Update", zap.Error(err))
			return err
		}
	}

	return dbTrx.GetSqlTx().Commit()
}

func (s *paymentConfigService) SetGatewayDisplayOrder(ctx context.Context, code string, order int) error {
	dbTrx := dao.NewTransactionPayment(ctx, s.log, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	if err := dbTrx.GetGatewayDAO().UpdateDisplayOrder(ctx, code, order); err != nil {
		s.log.Error(ctx, "paymentConfigService.SetGatewayDisplayOrder", zap.Error(err))
		return err
	}

	return dbTrx.GetSqlTx().Commit()
}

func (s *paymentConfigService) SetManualTransferDisplayOrder(ctx context.Context, order int) error {
	dbTrx := dao.NewTransactionPayment(ctx, s.log, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	setting, err := dbTrx.GetPaymentSettingDAO().GetByKey(ctx, app_payment.SettingKeyManualTransferEnabled)
	if err != nil {
		s.log.Error(ctx, "paymentConfigService.SetManualTransferDisplayOrder.GetByKey", zap.Error(err))
		return err
	}

	if setting == nil {
		if err := dbTrx.GetPaymentSettingDAO().Upsert(ctx, app_payment.SettingKeyManualTransferEnabled, false, order); err != nil {
			s.log.Error(ctx, "paymentConfigService.SetManualTransferDisplayOrder.Upsert", zap.Error(err))
			return err
		}
	} else {
		if err := dbTrx.GetPaymentSettingDAO().UpdateDisplayOrder(ctx, app_payment.SettingKeyManualTransferEnabled, order); err != nil {
			s.log.Error(ctx, "paymentConfigService.SetManualTransferDisplayOrder.Update", zap.Error(err))
			return err
		}
	}

	return dbTrx.GetSqlTx().Commit()
}
