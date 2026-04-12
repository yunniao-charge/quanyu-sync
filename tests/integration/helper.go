//go:build integration

package integration

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"quanyu-battery-sync/internal/config"
	"quanyu-battery-sync/internal/device"
	"quanyu-battery-sync/internal/quanyu"
	"quanyu-battery-sync/internal/storage"

	"go.uber.org/zap"
)

// TestHelper 集成测试辅助工具
type TestHelper struct {
	Config     *config.Config
	Storage    *storage.MongoStorage
	APIClient  *quanyu.Client
	Logger     *zap.Logger
	TestRunID  string
	cleanupFns []func()
}

// NewTestHelper 创建测试辅助实例
// configPath: 配置文件路径（相对于项目根目录）
func NewTestHelper(configPath string) (*TestHelper, error) {
	// 构造绝对路径
	if !filepath.IsAbs(configPath) {
		absPath, err := filepath.Abs(configPath)
		if err != nil {
			return nil, fmt.Errorf("获取配置文件绝对路径失败: %w", err)
		}
		configPath = absPath
	}

	// 加载配置
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %w", err)
	}

	// 创建测试用 logger
	logger := zap.NewNop()

	// 连接 MongoDB
	mongoStore, err := storage.NewMongoStorage(cfg.MongoDB, logger)
	if err != nil {
		return nil, fmt.Errorf("连接 MongoDB 失败: %w", err)
	}

	// 创建 API 客户端
	apiClient := quanyu.NewClient(cfg.Quanyu, logger)

	// 生成唯一的测试运行 ID
	testRunID := fmt.Sprintf("test_%d", time.Now().UnixNano())

	h := &TestHelper{
		Config:    cfg,
		Storage:   mongoStore,
		APIClient: apiClient,
		Logger:    logger,
		TestRunID: testRunID,
	}

	return h, nil
}

// GetTestUIDs 获取测试用的设备 UID 列表
func (h *TestHelper) GetTestUIDs(ctx context.Context, count int) ([]string, error) {
	provider := device.NewProvider(h.Config.DeviceAPI, h.Logger)
	if err := provider.Start(ctx); err != nil {
		return nil, fmt.Errorf("启动设备 Provider 失败: %w", err)
	}

	// 等待 UID 列表加载
	time.Sleep(2 * time.Second)

	uids := provider.GetUIDs()
	if len(uids) == 0 {
		return nil, fmt.Errorf("设备列表为空")
	}

	if len(uids) > count {
		uids = uids[:count]
	}

	return uids, nil
}

// Cleanup 清理测试数据
func (h *TestHelper) Cleanup() {
	for _, fn := range h.cleanupFns {
		fn()
	}
}

// AddCleanup 注册清理函数
func (h *TestHelper) AddCleanup(fn func()) {
	h.cleanupFns = append(h.cleanupFns, fn)
}

// Close 关闭所有连接
func (h *TestHelper) Close(ctx context.Context) {
	h.Cleanup()
	if h.Storage != nil {
		h.Storage.Close(ctx)
	}
	_ = h.Logger.Sync()
}
