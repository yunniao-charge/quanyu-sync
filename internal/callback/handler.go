package callback

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"quanyu-battery-sync/internal/storage"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"
)

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
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("读取回调 body 失败", zap.Error(err))
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	h.logger.Debug("收到回调", zap.String("body", string(body)))

	// 尝试判断数据类型并分发处理
	ctx := r.Context()

	// 尝试解析为各种类型
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		h.logger.Error("解析回调 JSON 失败", zap.Error(err))
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// 检查 data 字段来判断类型
	dataRaw, ok := raw["data"]
	if !ok {
		h.logger.Warn("回调数据缺少 data 字段")
		w.WriteHeader(http.StatusOK)
		return
	}

	// 尝试解析 data 数组中的第一个元素来判断类型
	var dataItems []map[string]json.RawMessage
	if err := json.Unmarshal(dataRaw, &dataItems); err != nil || len(dataItems) == 0 {
		h.logger.Warn("回调 data 为空或解析失败")
		w.WriteHeader(http.StatusOK)
		return
	}

	firstItem := dataItems[0]

	// 判断类型：有 event 字段 -> alarm; 有 remain 或 voltage -> info; 有 online 字段 -> online
	if _, hasEvent := firstItem["event"]; hasEvent {
		h.handleAlarm(ctx, body)
	} else if _, hasRemain := firstItem["remain"]; hasRemain {
		h.handleInfo(ctx, body)
	} else if _, hasOnline := firstItem["online"]; hasOnline {
		h.handleOnline(ctx, body)
	} else {
		h.logger.Warn("无法识别的回调数据类型")
	}

	w.WriteHeader(http.StatusOK)
}

// handleInfo 处理 info 推送
func (h *Handler) handleInfo(ctx context.Context, body []byte) {
	var payload InfoPushPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		h.logger.Error("解析 info 推送数据失败", zap.Error(err))
		return
	}

	h.logger.Info("收到 info 回调",
		zap.String("appid", payload.AppID),
		zap.Int("data_count", len(payload.Data)),
	)

	for _, item := range payload.Data {
		now := time.Now()

		// 存入 callback_info 集合
		infoDoc := &storage.CallbackInfoDoc{
			UID:        item.UID,
			DevType:    item.DevType,
			SN:         item.SN,
			Loc:        item.Loc,
			Remain:     item.Remain,
			Online:     item.Online,
			Voltage:    item.Voltage,
			Charge:     item.Charge,
			Discharge:  item.Discharge,
			BatTime:    item.BatTime,
			ReceivedAt: now,
			AppID:      payload.AppID,
		}

		if err := h.storage.UpsertCallbackInfo(ctx, infoDoc); err != nil {
			h.logger.Error("存储 info 回调数据失败",
				zap.String("uid", item.UID),
				zap.Error(err),
			)
			continue
		}

		// 同时更新 battery_details 快照
		fields := bson.D{
			{Key: "soc", Value: item.Remain},
			{Key: "online_status", Value: item.Online},
			{Key: "charge", Value: item.Charge},
			{Key: "discharge", Value: item.Discharge},
			{Key: "loc", Value: item.Loc},
			{Key: "bat_time", Value: item.BatTime},
		}
		if err := h.storage.UpdateBatteryDetailFromCallback(ctx, item.UID, fields); err != nil {
			h.logger.Error("从回调更新电池详情失败",
				zap.String("uid", item.UID),
				zap.Error(err),
			)
		}
	}
}

// handleAlarm 处理 alarm 推送
func (h *Handler) handleAlarm(ctx context.Context, body []byte) {
	var payload AlarmPushPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		h.logger.Error("解析 alarm 推送数据失败", zap.Error(err))
		return
	}

	h.logger.Info("收到 alarm 回调",
		zap.String("appid", payload.AppID),
		zap.Int("data_count", len(payload.Data)),
	)

	for _, item := range payload.Data {
		for _, event := range item.Events {
			alarmDoc := &storage.CallbackAlarmDoc{
				UID:        item.UID,
				Alarm:      event.Alarm,
				Type:       event.Type,
				Time:       event.Time,
				InfoObj:    item.InfoObj,
				ReceivedAt: time.Now(),
				AppID:      payload.AppID,
			}

			if err := h.storage.InsertCallbackAlarm(ctx, alarmDoc); err != nil {
				h.logger.Error("存储 alarm 回调数据失败",
					zap.String("uid", item.UID),
					zap.String("alarm", event.Alarm),
					zap.Error(err),
				)
			}
		}
	}
}

// handleOnline 处理 online 推送
func (h *Handler) handleOnline(ctx context.Context, body []byte) {
	var payload OnlinePushPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		h.logger.Error("解析 online 推送数据失败", zap.Error(err))
		return
	}

	h.logger.Info("收到 online 回调",
		zap.String("appid", payload.AppID),
		zap.Int("data_count", len(payload.Data)),
	)

	for _, item := range payload.Data {
		onlineDoc := &storage.CallbackOnlineDoc{
			UID:        item.UID,
			Online:     item.Online,
			Time:       item.Time,
			ReceivedAt: time.Now(),
			AppID:      payload.AppID,
		}

		if err := h.storage.InsertCallbackOnline(ctx, onlineDoc); err != nil {
			h.logger.Error("存储 online 回调数据失败",
				zap.String("uid", item.UID),
				zap.Error(err),
			)
		}
	}
}
