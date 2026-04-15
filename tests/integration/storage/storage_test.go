//go:build integration

package storage_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"quanyu-battery-sync/internal/storage"
	"quanyu-battery-sync/tests/integration"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var testHelper *integration.TestHelper

func TestMain(m *testing.M) {
	var err error
	testHelper, err = integration.NewTestHelper("../../../config.yaml")
	if err != nil {
		panic(fmt.Sprintf("创建 TestHelper 失败: %v", err))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := testHelper.Storage.EnsureIndexes(ctx); err != nil {
		panic(fmt.Sprintf("创建索引失败: %v", err))
	}

	m.Run()

	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	testHelper.Close(ctx2)
}

// ==================== battery_details ====================

func TestUpsertBatteryDetail_Insert(t *testing.T) {
	ctx := context.Background()
	uid := fmt.Sprintf("test_uid_%d", time.Now().UnixNano())

	doc := &storage.BatteryDetailDoc{
		UID:        uid,
		SOC:        80,
		Voltage:    "52.1",
		SyncSource: "pull",
		UpdatedAt:  time.Now(),
	}

	err := testHelper.Storage.UpsertBatteryDetail(ctx, doc)
	require.NoError(t, err)

	// 验证可以读取
	found, err := testHelper.Storage.GetBatteryDetail(ctx, uid)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, uid, found.UID)
	assert.Equal(t, 80, found.SOC)
	assert.Equal(t, "52.1", found.Voltage)

	// 清理
	cleanupDoc(t, "battery_details", uid)
}

func TestUpsertBatteryDetail_Update(t *testing.T) {
	ctx := context.Background()
	uid := fmt.Sprintf("test_uid_%d", time.Now().UnixNano())

	// 先插入
	doc1 := &storage.BatteryDetailDoc{
		UID:        uid,
		SOC:        50,
		Voltage:    "48.0",
		SyncSource: "pull",
		UpdatedAt:  time.Now(),
	}
	require.NoError(t, testHelper.Storage.UpsertBatteryDetail(ctx, doc1))

	// 再更新（同 UID）
	doc2 := &storage.BatteryDetailDoc{
		UID:        uid,
		SOC:        90,
		Voltage:    "53.2",
		SyncSource: "callback",
		UpdatedAt:  time.Now(),
	}
	require.NoError(t, testHelper.Storage.UpsertBatteryDetail(ctx, doc2))

	// 验证是更新而非插入
	found, err := testHelper.Storage.GetBatteryDetail(ctx, uid)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, 90, found.SOC)
	assert.Equal(t, "53.2", found.Voltage)
	assert.Equal(t, "callback", found.SyncSource)

	cleanupDoc(t, "battery_details", uid)
}

func TestGetBatteryDetail_NotFound(t *testing.T) {
	ctx := context.Background()
	found, err := testHelper.Storage.GetBatteryDetail(ctx, "nonexistent_uid")
	require.NoError(t, err)
	assert.Nil(t, found)
}

// ==================== battery_history ====================

func TestUpsertBatteryHistory(t *testing.T) {
	ctx := context.Background()
	uid := fmt.Sprintf("test_uid_%d", time.Now().UnixNano())
	now := time.Now()

	docs := []storage.BatteryHistoryDoc{
		{
			UID:       uid,
			Timestamp: now.Add(-2 * time.Minute),
			VVV:       "52.1",
			Cap:       "100",
			SyncedAt:  now,
		},
		{
			UID:       uid,
			Timestamp: now.Add(-1 * time.Minute),
			VVV:       "52.3",
			Cap:       "100",
			SyncedAt:  now,
		},
	}

	count, err := testHelper.Storage.UpsertBatteryHistory(ctx, docs)
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	// 相同数据再写入应去重（upsert）
	count2, err := testHelper.Storage.UpsertBatteryHistory(ctx, docs)
	require.NoError(t, err)
	assert.Equal(t, 2, count2) // modified, not newly inserted

	cleanupByUID(t, "battery_history", uid)
}

func TestUpsertBatteryHistory_Empty(t *testing.T) {
	ctx := context.Background()
	count, err := testHelper.Storage.UpsertBatteryHistory(ctx, nil)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

// ==================== battery_traces ====================

func TestUpsertBatteryTraces(t *testing.T) {
	ctx := context.Background()
	uid := fmt.Sprintf("test_uid_%d", time.Now().UnixNano())
	now := time.Now()

	docs := []storage.BatteryTraceDoc{
		{UID: uid, Loc: "116.404,39.915", LocTime: now.Add(-5 * time.Minute).Format("20060102150405"), SyncedAt: now},
		{UID: uid, Loc: "116.405,39.916", LocTime: now.Add(-3 * time.Minute).Format("20060102150405"), SyncedAt: now},
	}

	count, err := testHelper.Storage.UpsertBatteryTraces(ctx, docs)
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	cleanupByUID(t, "battery_traces", uid)
}

// ==================== battery_events ====================

func TestUpsertBatteryEvents(t *testing.T) {
	ctx := context.Background()
	uid := fmt.Sprintf("test_uid_%d", time.Now().UnixNano())
	now := time.Now()

	docs := []storage.BatteryEventDoc{
		{UID: uid, Alarm: "overvoltage", Time: now.Add(-2 * time.Minute).Format("2006-01-02 15:04:05"), V1: "val1", SyncedAt: now},
		{UID: uid, Alarm: "undervoltage", Time: now.Add(-1 * time.Minute).Format("2006-01-02 15:04:05"), V1: "val2", SyncedAt: now},
	}

	count, err := testHelper.Storage.UpsertBatteryEvents(ctx, docs)
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	cleanupByUID(t, "battery_events", uid)
}

// ==================== charge_records ====================

func TestUpsertChargeRecords(t *testing.T) {
	ctx := context.Background()
	uid := fmt.Sprintf("test_uid_%d", time.Now().UnixNano())
	now := time.Now()

	docs := []storage.ChargeRecordDoc{
		{UID: uid, DeviceID: "dev1", ChargeBegin: now.Add(-1 * time.Hour).Format(time.RFC3339), IDXAuto: 1001, SyncedAt: now},
		{UID: uid, DeviceID: "dev2", ChargeBegin: now.Add(-30 * time.Minute).Format(time.RFC3339), IDXAuto: 1002, SyncedAt: now},
	}

	count, err := testHelper.Storage.UpsertChargeRecords(ctx, docs)
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	// 去重
	docs[0].ChargeBegin = "updated_time"
	count2, err := testHelper.Storage.UpsertChargeRecords(ctx, docs)
	require.NoError(t, err)
	assert.Equal(t, 2, count2)

	cleanupByUID(t, "charge_records", uid)
}

// ==================== callback_alarms ====================

func TestInsertCallbackAlarm(t *testing.T) {
	ctx := context.Background()
	uid := fmt.Sprintf("test_uid_%d", time.Now().UnixNano())

	doc := &storage.CallbackAlarmDoc{
		UID:        uid,
		Alarm:      "overvoltage",
		Type:       1,
		Time:       time.Now().Format("2006-01-02 15:04:05"),
		ReceivedAt: time.Now(),
		AppID:      "test_app",
	}

	require.NoError(t, testHelper.Storage.InsertCallbackAlarm(ctx, doc))

	cleanupByUID(t, "callback_alarms", uid)
}

// ==================== callback_online ====================

func TestInsertCallbackOnline(t *testing.T) {
	ctx := context.Background()
	uid := fmt.Sprintf("test_uid_%d", time.Now().UnixNano())

	doc := &storage.CallbackOnlineDoc{
		UID:        uid,
		Online:     1,
		Time:       time.Now().Format("2006-01-02 15:04:05"),
		ReceivedAt: time.Now(),
		AppID:      "test_app",
	}

	require.NoError(t, testHelper.Storage.InsertCallbackOnline(ctx, doc))

	cleanupByUID(t, "callback_online", uid)
}

// ==================== sync_states ====================

func TestSyncState_CreateAndUpdate(t *testing.T) {
	ctx := context.Background()
	uid := fmt.Sprintf("test_uid_%d", time.Now().UnixNano())
	syncType := "detail"

	// 首次读取应为空
	state, err := testHelper.Storage.GetSyncState(ctx, uid, syncType)
	require.NoError(t, err)
	assert.Nil(t, state)

	// 更新同步时间（自动创建）
	syncTime := time.Now().Format("2006-01-02 15:04:05")
	require.NoError(t, testHelper.Storage.UpdateSyncTime(ctx, uid, syncType, syncTime))

	// 验证状态已创建
	state, err = testHelper.Storage.GetSyncState(ctx, uid, syncType)
	require.NoError(t, err)
	require.NotNil(t, state)
	assert.Equal(t, uid, state.UID)
	assert.Equal(t, syncType, state.SyncType)
	assert.Equal(t, syncTime, state.LastSyncTime)
	assert.Equal(t, 1, state.SyncCount)

	// 再次更新同步时间
	syncTime2 := time.Now().Format("2006-01-02 15:04:05")
	require.NoError(t, testHelper.Storage.UpdateSyncTime(ctx, uid, syncType, syncTime2))

	state, err = testHelper.Storage.GetSyncState(ctx, uid, syncType)
	require.NoError(t, err)
	require.NotNil(t, state)
	assert.Equal(t, syncTime2, state.LastSyncTime)
	assert.Equal(t, 2, state.SyncCount)

	cleanupSyncState(t, uid, syncType)
}

func TestSyncState_UpdateError(t *testing.T) {
	ctx := context.Background()
	uid := fmt.Sprintf("test_uid_%d", time.Now().UnixNano())
	syncType := "detail"

	require.NoError(t, testHelper.Storage.UpdateSyncError(ctx, uid, syncType, "connection timeout"))

	state, err := testHelper.Storage.GetSyncState(ctx, uid, syncType)
	require.NoError(t, err)
	require.NotNil(t, state)
	assert.Equal(t, "connection timeout", state.LastError)
	assert.Equal(t, 1, state.ErrorCount)

	// 再报一次错误
	require.NoError(t, testHelper.Storage.UpdateSyncError(ctx, uid, syncType, "network error"))

	state, err = testHelper.Storage.GetSyncState(ctx, uid, syncType)
	require.NoError(t, err)
	assert.Equal(t, 2, state.ErrorCount)
	assert.Equal(t, "network error", state.LastError)

	cleanupSyncState(t, uid, syncType)
}

func TestSyncState_UniqueByUIDAndType(t *testing.T) {
	ctx := context.Background()
	uid := fmt.Sprintf("test_uid_%d", time.Now().UnixNano())

	// 同 UID 不同 type 应是不同记录
	require.NoError(t, testHelper.Storage.UpdateSyncTime(ctx, uid, "detail", "2026-01-01 00:00:00"))
	require.NoError(t, testHelper.Storage.UpdateSyncTime(ctx, uid, "history_data", "2026-01-01 01:00:00"))

	state1, _ := testHelper.Storage.GetSyncState(ctx, uid, "detail")
	state2, _ := testHelper.Storage.GetSyncState(ctx, uid, "history_data")

	require.NotNil(t, state1)
	require.NotNil(t, state2)
	assert.NotEqual(t, state1.LastSyncTime, state2.LastSyncTime)

	cleanupSyncState(t, uid, "detail")
	cleanupSyncState(t, uid, "history_data")
}

// ==================== UpdateBatteryDetailFromCallback ====================

func TestUpdateBatteryDetailFromCallback(t *testing.T) {
	ctx := context.Background()
	uid := fmt.Sprintf("test_uid_%d", time.Now().UnixNano())

	// 先创建一条记录
	doc := &storage.BatteryDetailDoc{
		UID:        uid,
		SOC:        50,
		SyncSource: "pull",
		UpdatedAt:  time.Now(),
	}
	require.NoError(t, testHelper.Storage.UpsertBatteryDetail(ctx, doc))

	// 通过回调更新部分字段
	fields := bson.D{
		{Key: "soc", Value: 85},
		{Key: "online_status", Value: 1},
		{Key: "charge", Value: 1},
	}
	require.NoError(t, testHelper.Storage.UpdateBatteryDetailFromCallback(ctx, uid, fields))

	// 验证更新
	found, err := testHelper.Storage.GetBatteryDetail(ctx, uid)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, 85, found.SOC)
	assert.Equal(t, "callback", found.SyncSource)

	cleanupDoc(t, "battery_details", uid)
}

// ==================== helpers ====================

func cleanupDoc(t *testing.T, collection, uid string) {
	t.Helper()
	ctx := context.Background()
	col := testHelper.Storage.Collection(collection)
	_, err := col.DeleteMany(ctx, bson.D{{Key: "uid", Value: uid}})
	if err != nil {
		t.Logf("清理 %s/%s 失败: %v", collection, uid, err)
	}
}

func cleanupByUID(t *testing.T, collection, uid string) {
	t.Helper()
	cleanupDoc(t, collection, uid)
}

func cleanupSyncState(t *testing.T, uid, syncType string) {
	t.Helper()
	ctx := context.Background()
	col := testHelper.Storage.Collection("sync_states")
	_, err := col.DeleteMany(ctx, bson.D{
		{Key: "uid", Value: uid},
		{Key: "sync_type", Value: syncType},
	})
	if err != nil {
		t.Logf("清理 sync_states/%s/%s 失败: %v", uid, syncType, err)
	}
}
