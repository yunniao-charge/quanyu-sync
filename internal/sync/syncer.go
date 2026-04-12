package sync

import (
	"context"

	"quanyu-battery-sync/internal/config"
	"quanyu-battery-sync/internal/device"
	"quanyu-battery-sync/internal/quanyu"
	"quanyu-battery-sync/internal/storage"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// Syncer 同步调度器
type Syncer struct {
	client   quanyu.QuanyuClient
	storage  storage.SyncStorage
	provider device.DeviceProvider
	config   config.SyncConfig
	logger   *zap.Logger
	cron     *cron.Cron
}

// NewSyncer 创建同步调度器
func NewSyncer(
	client quanyu.QuanyuClient,
	store storage.SyncStorage,
	provider device.DeviceProvider,
	cfg config.SyncConfig,
	logger *zap.Logger,
) *Syncer {
	return &Syncer{
		client:   client,
		storage:  store,
		provider: provider,
		config:   cfg,
		logger:   logger,
		cron:     cron.New(cron.WithSeconds()),
	}
}

// Start 启动同步调度器
func (s *Syncer) Start(ctx context.Context) error {
	// 注册各数据类型的同步任务
	tasks := []struct {
		name    string
		cfg     config.SyncTaskConfig
		fn      func(context.Context, string) error
	}{
		{"detail", s.config.Detail, s.syncDetailTask},
		{"history_data", s.config.HistoryData, s.syncHistoryDataTask},
		{"trace", s.config.Trace, s.syncTraceTask},
		{"event", s.config.Event, s.syncEventTask},
		{"charge_data", s.config.ChargeData, s.syncChargeDataTask},
	}

	for _, task := range tasks {
		if !task.cfg.Enabled {
			s.logger.Info("同步任务已禁用", zap.String("type", task.name))
			continue
		}

		taskFn := task.fn // capture
		taskName := task.name
		_, err := s.cron.AddFunc(task.cfg.Cron, func() {
			s.runSyncRound(context.Background(), taskName, taskFn)
		})
		if err != nil {
			s.logger.Error("注册 cron 任务失败",
				zap.String("type", task.name),
				zap.String("cron", task.cfg.Cron),
				zap.Error(err),
			)
			continue
		}
		s.logger.Info("同步任务已注册",
			zap.String("type", task.name),
			zap.String("cron", task.cfg.Cron),
		)
	}

	s.cron.Start()
	return nil
}

// Stop 停止同步调度器
func (s *Syncer) Stop() {
	if s.cron != nil {
		s.cron.Stop()
	}
}

// runSyncRound 执行一轮同步
func (s *Syncer) runSyncRound(ctx context.Context, syncType string, fn func(context.Context, string) error) {
	uids := s.provider.GetUIDs()
	if len(uids) == 0 {
		s.logger.Warn("设备列表为空，跳过同步", zap.String("type", syncType))
		return
	}

	s.logger.Info("开始同步轮次",
		zap.String("type", syncType),
		zap.Int("uid_count", len(uids)),
	)

	successCount := 0
	failCount := 0

	for _, uid := range uids {
		if err := fn(ctx, uid); err != nil {
			failCount++
			s.logger.Error("同步失败",
				zap.String("type", syncType),
				zap.String("uid", uid),
				zap.Error(err),
			)
		} else {
			successCount++
		}
	}

	s.logger.Info("同步轮次完成",
		zap.String("type", syncType),
		zap.Int("success", successCount),
		zap.Int("failed", failCount),
		zap.Int("total", len(uids)),
	)
}
