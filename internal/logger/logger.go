package logger

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	AppLogger   *zap.Logger
	Sugar       *zap.SugaredLogger
	APIDebugLog bool
	debugLogger *zap.Logger
)

func Init(dir, level string, maxSize, maxBackups, maxAge int, apiDebugLog bool) error {
	APIDebugLog = apiDebugLog

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 解析日志级别
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		zapLevel = zapcore.InfoLevel
	}

	// 主日志 encoder
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "ts"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder

	// console encoder
	consoleEncoder := zapcore.NewConsoleEncoder(encoderCfg)

	// file encoder
	fileEncoder := zapcore.NewJSONEncoder(encoderCfg)

	// 主日志文件 writer
	fileWriter := &lumberjack.Logger{
		Filename:   filepath.Join(dir, "app.log"),
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
	}

	// console + file
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapLevel),
		zapcore.NewCore(fileEncoder, zapcore.AddSync(fileWriter), zapLevel),
	)

	AppLogger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	Sugar = AppLogger.Sugar()

	// 同步统计日志 - 单独文件
	syncWriter := &lumberjack.Logger{
		Filename:   filepath.Join(dir, "sync.log"),
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
	}
	syncCore := zapcore.NewCore(fileEncoder, zapcore.AddSync(syncWriter), zapLevel)
	_ = syncCore // syncLogger 在 Syncer 内部使用

	// API 调试日志
	if apiDebugLog {
		debugDir := filepath.Join(dir, "debug")
		if err := os.MkdirAll(debugDir, 0755); err != nil {
			return err
		}
		debugWriter := &lumberjack.Logger{
			Filename:   filepath.Join(debugDir, "api.log"),
			MaxSize:    maxSize,
			MaxBackups: maxBackups,
			MaxAge:     maxAge,
		}
		debugCore := zapcore.NewCore(fileEncoder, zapcore.AddSync(debugWriter), zapcore.DebugLevel)
		debugLogger = zap.New(debugCore)
	}

	return nil
}

// LogAPI 记录 API 请求/响应到调试日志
func LogAPI(endpoint, uid string, attempt int, request, response string) {
	if debugLogger != nil {
		debugLogger.Info("API debug",
			zap.String("endpoint", endpoint),
			zap.String("uid", uid),
			zap.Int("attempt", attempt),
			zap.String("request", request),
			zap.String("response", response),
		)
	}
}

// SyncLogger 返回同步统计专用 logger
func SyncLogger(dir string, maxSize, maxBackups, maxAge int) *zap.Logger {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "ts"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder
	fileEncoder := zapcore.NewJSONEncoder(encoderCfg)

	syncWriter := &lumberjack.Logger{
		Filename:   filepath.Join(dir, "sync.log"),
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
	}

	return zap.New(zapcore.NewCore(fileEncoder, zapcore.AddSync(syncWriter), zapcore.InfoLevel))
}

func Sync() {
	if AppLogger != nil {
		_ = AppLogger.Sync()
	}
	if debugLogger != nil {
		_ = debugLogger.Sync()
	}
}
