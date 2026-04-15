package callback

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"quanyu-battery-sync/internal/logger"
	"quanyu-battery-sync/internal/storage"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"
)

// formatBatTime 将回调的紧凑时间格式 "20060415095539" 统一为 "2006-01-02 15:04:05"
// 如果已经是标准格式或为空，原样返回
func formatBatTime(s string) string {
	if s == "" || len(s) != 14 {
		return s
	}
	t, err := time.Parse("20060102150405", s)
	if err != nil {
		return s
	}
	return t.Format("2006-01-02 15:04:05")
}

// Handler HTTP 回调处理器
type Handler struct {
	storage storage.CallbackStorage
	logger  *zap.Logger
}

// NewHandler 创建回调处理器
func NewHandler(store storage.CallbackStorage, logger *zap.Logger) *Handler {
	return &Handler{
		storage: store,
		logger:  logger,
	}
}

// ServeHTTP 统一回调入口
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.logger.Warn("[callback] 收到非 POST 请求",
			zap.String("method", r.Method),
			zap.String("remote_addr", r.RemoteAddr),
			zap.String("path", r.URL.Path),
		)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("[callback] 读取 body 失败", zap.Error(err))
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// 回调调试日志：记录完整请求
	logger.LogCallbackDebug("[callback] 收到 HTTP 请求",
		zap.String("remote_addr", r.RemoteAddr),
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.String("content_type", r.Header.Get("Content-Type")),
		zap.Int("body_length", len(body)),
		zap.String("body", string(body)),
	)

	ctx := r.Context()

	// 解析顶层结构，用 type 字段分发
	var raw struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		h.logger.Error("[callback] 解析 JSON 失败",
			zap.Error(err),
			zap.String("body", string(body)),
		)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	switch raw.Type {
	case "info":
		h.handleInfo(ctx, raw.Data)
	case "alarm":
		h.handleAlarm(ctx, raw.Data)
	case "online":
		h.handleOnline(ctx, raw.Data)
	default:
		h.logger.Warn("[callback] 未知类型",
			zap.String("type", raw.Type),
			zap.String("body", string(body)),
		)
	}

	w.WriteHeader(http.StatusOK)
}

// handleInfo 处理 info 推送（已验证格式，正常解析存储）
func (h *Handler) handleInfo(ctx context.Context, dataRaw json.RawMessage) {
	var item InfoPushItem
	if err := json.Unmarshal(dataRaw, &item); err != nil {
		h.logger.Error("[callback:info] 解析失败", zap.Error(err))
		return
	}

	h.logger.Debug("[callback:info] 收到",
		zap.String("uid", item.UID),
		zap.Int("remain", item.Remain),
		zap.Int("online", item.Online),
		zap.String("voltage", item.Voltage),
		zap.String("loc", item.Loc),
	)

	// 统一 bat_time 格式为 "2006-01-02 15:04:05"
	batTime := formatBatTime(item.BatTime)

	fields := bson.D{
		{Key: "online_status", Value: item.Online},
		{Key: "charge", Value: item.Charge},
		{Key: "discharge", Value: item.Discharge},
		{Key: "loc", Value: item.Loc},
		{Key: "bat_time", Value: batTime},
	}
	if err := h.storage.UpdateBatteryDetailFromCallback(ctx, item.UID, fields); err != nil {
		h.logger.Error("[callback:info] 更新 battery_details 失败",
			zap.String("uid", item.UID),
			zap.Error(err),
		)
	}
}

// handleAlarm 处理 alarm 推送（已验证，2026-04-15）
func (h *Handler) handleAlarm(ctx context.Context, dataRaw json.RawMessage) {
	var item AlarmPushItem
	if err := json.Unmarshal(dataRaw, &item); err != nil {
		h.logger.Error("[callback:alarm] 解析失败", zap.Error(err))
		return
	}

	h.logger.Debug("[callback:alarm] 收到",
		zap.String("uid", item.UID),
		zap.String("alarmCode", item.AlarmCode),
		zap.String("time", item.Time),
	)

	now := time.Now()

	doc := &storage.CallbackAlarmDoc{
		UID:        item.UID,
		Alarm:      item.AlarmCode,
		AlarmData:  item.AlarmData,
		Time:       item.Time,
		ReceivedAt: now,
		AppID:      "callback",
	}

	if err := h.storage.InsertCallbackAlarm(ctx, doc); err != nil {
		h.logger.Error("[callback:alarm] 存储失败",
			zap.String("uid", item.UID),
			zap.Error(err),
		)
	}
}

// handleOnline 处理 online 推送（已验证，2026-04-15）
func (h *Handler) handleOnline(ctx context.Context, dataRaw json.RawMessage) {
	var item OnlinePushItem
	if err := json.Unmarshal(dataRaw, &item); err != nil {
		h.logger.Error("[callback:online] 解析失败", zap.Error(err))
		return
	}

	// 毫秒时间戳转格式化字符串
	t := time.UnixMilli(item.Time)
	timeStr := t.Format("20060102150405")

	h.logger.Debug("[callback:online] 收到",
		zap.String("uid", item.UID),
		zap.Int("online", item.Online),
		zap.String("time", timeStr),
	)

	now := time.Now()

	doc := &storage.CallbackOnlineDoc{
		UID:        item.UID,
		Online:     item.Online,
		Time:       timeStr,
		ReceivedAt: now,
		AppID:      "callback",
	}

	if err := h.storage.InsertCallbackOnline(ctx, doc); err != nil {
		h.logger.Error("[callback:online] 存储失败",
			zap.String("uid", item.UID),
			zap.Error(err),
		)
		return
	}

	// 更新 battery_details 在线状态
	fields := bson.D{
		{Key: "online_status", Value: item.Online},
	}
	if err := h.storage.UpdateBatteryDetailFromCallback(ctx, item.UID, fields); err != nil {
		h.logger.Error("[callback:online] 更新 battery_details 失败",
			zap.String("uid", item.UID),
			zap.Error(err),
		)
	}
}
