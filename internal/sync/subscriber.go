package sync

import (
	"context"
	"fmt"
	"time"

	"quanyu-battery-sync/internal/config"
	"quanyu-battery-sync/internal/device"
	"quanyu-battery-sync/internal/logger"
	"quanyu-battery-sync/internal/quanyu"

	"go.uber.org/zap"
)

// Subscriber 订阅管理器
type Subscriber struct {
	client   quanyu.QuanyuClient
	provider device.DeviceProvider
	config   config.SubscribeConfig
	callback config.CallbackConfig
	logger   *zap.Logger
}

// NewSubscriber 创建订阅管理器
func NewSubscriber(
	client quanyu.QuanyuClient,
	provider device.DeviceProvider,
	subCfg config.SubscribeConfig,
	cbCfg config.CallbackConfig,
	logger *zap.Logger,
) *Subscriber {
	return &Subscriber{
		client:   client,
		provider: provider,
		config:   subCfg,
		callback: cbCfg,
		logger:   logger,
	}
}

// StartRenewLoop 启动订阅续订循环
func (sub *Subscriber) StartRenewLoop(ctx context.Context) {
	sub.logger.Info("订阅续订循环启动",
		zap.Duration("interval", sub.config.RenewInterval),
		zap.Int("batch_size", sub.config.BatchSize),
		zap.String("notify_url", sub.callback.NotifyURL),
	)

	// 首次立即订阅
	if err := sub.subscribeAll(ctx); err != nil {
		sub.logger.Error("初始订阅失败", zap.Error(err))
	}

	go func() {
		ticker := time.NewTicker(sub.config.RenewInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				sub.logger.Info("订阅续订循环停止")
				return
			case <-ticker.C:
				if err := sub.subscribeAll(ctx); err != nil {
					sub.logger.Error("订阅续订失败", zap.Error(err))
				}
			}
		}
	}()
}

// subscribeAll 分批订阅所有设备
func (sub *Subscriber) subscribeAll(ctx context.Context) error {
	uids := sub.provider.GetUIDs()
	if len(uids) == 0 {
		sub.logger.Warn("设备列表为空，跳过订阅")
		return nil
	}

	notifyURL := sub.callback.NotifyURL
	batchSize := sub.config.BatchSize

	logger.LogCallbackDebug("[subscribe] 开始订阅",
		zap.Int("total_devices", len(uids)),
		zap.Int("batch_size", batchSize),
		zap.String("notify_url", notifyURL),
		zap.Strings("sub_data", sub.config.SubData),
	)

	successCount := 0
	failCount := 0

	for i := 0; i < len(uids); i += batchSize {
		end := i + batchSize
		if end > len(uids) {
			end = len(uids)
		}

		batch := uids[i:end]

		logger.LogCallbackDebug("[subscribe] 批次",
			zap.Int("batch_index", i/batchSize+1),
			zap.Int("batch_size", len(batch)),
			zap.String("first_uid", batch[0]),
			zap.String("last_uid", batch[len(batch)-1]),
			zap.String("notify_url", notifyURL),
		)

		// 尝试批量订阅：用第一个 uid 签名，list 传整个 batch
		resp, err := sub.client.SubscribeV2(ctx, batch[0], batch, sub.config.SubData, notifyURL)
		if err != nil || (resp != nil && resp.Errno != 0) {
			batchErr := err
			if resp != nil && resp.Errno != 0 {
				batchErr = fmt.Errorf("errno=%d: %s", resp.Errno, resp.Errmsg)
			}
			sub.logger.Warn("批量订阅失败，降级为逐个订阅",
				zap.Int("batch_size", len(batch)),
				zap.Error(batchErr),
			)
			// 降级为逐个订阅
			for _, uid := range batch {
				resp, err := sub.client.SubscribeV2(ctx, uid, []string{uid}, sub.config.SubData, notifyURL)
				if err != nil {
					failCount++
					sub.logger.Warn("订阅设备失败",
						zap.String("uid", uid),
						zap.Error(err),
					)
					continue
				}

				if resp.Errno != 0 {
					failCount++
					sub.logger.Warn("订阅设备返回错误",
						zap.String("uid", uid),
						zap.Int("errno", resp.Errno),
						zap.String("errmsg", resp.Errmsg),
					)
					continue
				}

				successCount++
			}
			continue
		}

		successCount += len(batch)
	}

	sub.logger.Info("订阅完成",
		zap.Int("total", len(uids)),
		zap.Int("success", successCount),
		zap.Int("failed", failCount),
	)

	logger.LogCallbackDebug("[subscribe] 订阅完成",
		zap.Int("total", len(uids)),
		zap.Int("success", successCount),
		zap.Int("failed", failCount),
		zap.String("notify_url", notifyURL),
	)

	return nil
}
