package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"quanyu-battery-sync/internal/callback"
	"quanyu-battery-sync/internal/config"
	"quanyu-battery-sync/internal/device"
	"quanyu-battery-sync/internal/logger"
	"quanyu-battery-sync/internal/quanyu"
	"quanyu-battery-sync/internal/storage"
	syncpkg "quanyu-battery-sync/internal/sync"
)

func main() {
	configPath := flag.String("config", "config.yaml", "配置文件路径")
	flag.Parse()

	// 1. 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 2. 初始化日志
	if err := logger.Init(cfg.Log.Dir, cfg.Log.Level, cfg.Log.MaxSize, cfg.Log.MaxBackups, cfg.Log.MaxAge, cfg.Log.APIDebugLog, cfg.Log.CallbackDebugLog); err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	log := logger.Sugar
	log.Info("========================================")
	log.Info("全裕电池数据同步服务启动")
	log.Info("========================================")
	log.Infow("配置",
		"api_base_url", cfg.Quanyu.BaseURL,
		"device_api_url", cfg.DeviceAPI.URL,
		"mongodb_uri", cfg.MongoDB.URI,
		"mongodb_db", cfg.MongoDB.Database,
		"callback_addr", cfg.Callback.ListenAddr,
	)

	// 3. 连接 MongoDB
	ctx := context.Background()
	mongoStore, err := storage.NewMongoStorage(cfg.MongoDB, logger.AppLogger)
	if err != nil {
		log.Fatalw("连接 MongoDB 失败", "error", err)
	}

	// 4. 确保索引
	if err := mongoStore.EnsureIndexes(ctx); err != nil {
		log.Fatalw("创建索引失败", "error", err)
	}
	log.Info("MongoDB 索引创建完成")

	// 5. 创建全裕 API 客户端
	apiClient := quanyu.NewClient(cfg.Quanyu, logger.AppLogger)

	// 6. 创建设备列表 Provider
	provider := device.NewProvider(cfg.DeviceAPI, logger.AppLogger)

	// 7. 启动回调 HTTP 服务
	cbHandler := callback.NewHandler(mongoStore, logger.AppLogger)
	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/callback/push", cbHandler.ServeHTTP)
	httpMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	httpServer := &http.Server{
		Addr:    cfg.Callback.ListenAddr,
		Handler: httpMux,
	}

	go func() {
		log.Infow("回调 HTTP 服务启动", "addr", cfg.Callback.ListenAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalw("HTTP 服务异常", "error", err)
		}
	}()

	// 8. 创建并启动同步调度器
	syncer := syncpkg.NewSyncer(apiClient, mongoStore, provider, cfg.Sync, logger.AppLogger)

	// 启动 Provider
	if err := provider.Start(ctx); err != nil {
		log.Warnw("启动设备 Provider 失败", "error", err)
	}

	// 启动同步调度器
	if err := syncer.Start(ctx); err != nil {
		log.Fatalw("启动同步调度器失败", "error", err)
	}

	// 9. 创建并启动订阅续订
	subscriber := syncpkg.NewSubscriber(apiClient, provider, cfg.Subscribe, cfg.Callback, logger.AppLogger)
	subscriber.StartRenewLoop(ctx)

	log.Info("所有服务已启动，等待退出信号...")

	// 10. 等待退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("服务正在停止...")

	// 优雅退出
	syncer.Stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Errorw("HTTP 服务关闭异常", "error", err)
	}

	if err := mongoStore.Close(shutdownCtx); err != nil {
		log.Errorw("MongoDB 关闭异常", "error", err)
	}

	log.Info("服务已停止")
}
