package device

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"quanyu-battery-sync/internal/config"

	"go.uber.org/zap"
)

// Provider 设备 UID 列表提供者
type Provider struct {
	apiURL     string
	httpClient *http.Client
	interval   time.Duration
	mu         sync.RWMutex
	uids       []string
	logger     *zap.Logger
}

// NewProvider 创建设备列表 Provider
func NewProvider(cfg config.DeviceAPIConfig, logger *zap.Logger) *Provider {
	return &Provider{
		apiURL: cfg.URL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		interval: cfg.RefreshInterval,
		logger:   logger,
	}
}

// Start 启动定时刷新
func (p *Provider) Start(ctx context.Context) error {
	// 首次加载
	if err := p.refreshUIDs(ctx); err != nil {
		p.logger.Warn("初始加载设备列表失败，下次重试", zap.Error(err))
	}

	go p.refreshLoop(ctx)
	return nil
}

// GetUIDs 获取当前 uid 列表（线程安全）
func (p *Provider) GetUIDs() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make([]string, len(p.uids))
	copy(result, p.uids)
	return result
}

func (p *Provider) refreshLoop(ctx context.Context) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := p.refreshUIDs(ctx); err != nil {
				p.logger.Error("刷新设备列表失败", zap.Error(err))
			}
		}
	}
}

func (p *Provider) refreshUIDs(ctx context.Context) error {
	p.logger.Info("开始刷新设备列表", zap.String("url", p.apiURL))

	req, err := http.NewRequestWithContext(ctx, "GET", p.apiURL, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("请求设备列表API失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("设备列表API返回非200状态: %d", resp.StatusCode)
	}

	var uids []string
	if err := json.NewDecoder(resp.Body).Decode(&uids); err != nil {
		return fmt.Errorf("解析设备列表失败: %w", err)
	}

	p.mu.Lock()
	p.uids = uids
	p.mu.Unlock()

	p.logger.Info("设备列表刷新完成", zap.Int("count", len(uids)))
	return nil
}
