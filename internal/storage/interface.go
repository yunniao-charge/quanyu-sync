package storage

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// SyncStorage 同步数据存储接口（供 Syncer 使用）
type SyncStorage interface {
	// Battery detail 快照
	UpsertBatteryDetail(ctx context.Context, doc *BatteryDetailDoc) error
	GetBatteryDetail(ctx context.Context, uid string) (*BatteryDetailDoc, error)

	// 历史时序数据
	UpsertBatteryHistory(ctx context.Context, docs []BatteryHistoryDoc) (int, error)

	// 轨迹数据
	UpsertBatteryTraces(ctx context.Context, docs []BatteryTraceDoc) (int, error)

	// 事件数据
	UpsertBatteryEvents(ctx context.Context, docs []BatteryEventDoc) (int, error)

	// 充放电记录
	UpsertChargeRecords(ctx context.Context, docs []ChargeRecordDoc) (int, error)

	// 同步状态管理
	GetSyncState(ctx context.Context, uid, syncType string) (*SyncStateDoc, error)
	UpdateSyncTime(ctx context.Context, uid, syncType, syncTime string) error
	UpdateSyncError(ctx context.Context, uid, syncType, errMsg string) error
}

// CallbackStorage 回调数据存储接口（供 Callback Handler 使用）
type CallbackStorage interface {
	UpsertCallbackInfo(ctx context.Context, doc *CallbackInfoDoc) error
	InsertCallbackAlarm(ctx context.Context, doc *CallbackAlarmDoc) error
	InsertCallbackOnline(ctx context.Context, doc *CallbackOnlineDoc) error
	UpdateBatteryDetailFromCallback(ctx context.Context, uid string, fields bson.D) error
}

// 编译时检查 *MongoStorage 实现了两个接口
var (
	_ SyncStorage     = (*MongoStorage)(nil)
	_ CallbackStorage = (*MongoStorage)(nil)
)
