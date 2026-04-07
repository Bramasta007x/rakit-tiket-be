package cron

import (
	"context"
	"fmt"

	"rakit-tiket-be/internal/app/app_order/service"
	"rakit-tiket-be/pkg/util"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type Scheduler struct {
	cron         *cron.Cron
	orderService service.OrderService
	log          util.LogUtil
}

func NewScheduler(orderService service.OrderService, log util.LogUtil) *Scheduler {
	return &Scheduler{
		cron:         cron.New(cron.WithSeconds()),
		orderService: orderService,
		log:          log,
	}
}

func (s *Scheduler) Start() error {
	_, err := s.cron.AddFunc("0 */5 * * * *", s.cleanupExpiredOrders)
	if err != nil {
		return fmt.Errorf("failed to add expired orders cleanup cron: %w", err)
	}

	s.cron.Start()
	//s.log.Info(context.Background(), "Cron scheduler started", zap.Strings("jobs", []string{"cleanupExpiredOrders (every 5 minutes)"}))
	return nil
}

func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
	//s.log.Info(context.Background(), "Cron scheduler stopped")
}

func (s *Scheduler) cleanupExpiredOrders() {
	ctx := context.Background()
	//s.log.Info(ctx, "Running expired orders cleanup job")

	count, err := s.orderService.UpdateExpiredOrders(ctx)
	if err != nil {
		s.log.Error(ctx, "Failed to cleanup expired orders", zap.Error(err))
		return
	}

	if count > 0 {
		//s.log.Info(ctx, "Expired orders cleanup completed", zap.Int64("updated", count))
	}
}
