//go:build integration

package callback_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"quanyu-battery-sync/internal/callback"
	"quanyu-battery-sync/internal/storage"
	"quanyu-battery-sync/tests/integration"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"
)

var testHelper *integration.TestHelper
var handler *callback.Handler

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

	handler = callback.NewHandler(testHelper.Storage, zap.NewNop())

	m.Run()

	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	testHelper.Close(ctx2)
}

// TestInfoCallback 验证 info 推送正确存入 callback_info 且 battery_details 快照更新
func TestInfoCallback(t *testing.T) {
	uid := fmt.Sprintf("test_uid_%d", time.Now().UnixNano())
	defer cleanupByUID(t, "callback_info", uid)
	defer cleanupByUID(t, "battery_details", uid)

	payload := callback.InfoPushPayload{
		AppID: "test_app",
		Data: []callback.InfoPushItem{
			{
				UID:       uid,
				DevType:   1,
				SN:        "SN123",
				Loc:       "116.404,39.915",
				Remain:    85,
				Online:    1,
				Voltage:   520,
				Charge:    1,
				Discharge: 0,
				BatTime:   "2026-01-01 12:00:00",
			},
		},
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/callback/push", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// 验证 callback_info 数据
	ctx := context.Background()
	col := testHelper.Storage.Collection("callback_info")
	var infoDoc storage.CallbackInfoDoc
	err := col.FindOne(ctx, bson.D{{Key: "uid", Value: uid}}).Decode(&infoDoc)
	require.NoError(t, err, "callback_info 应有数据")
	assert.Equal(t, uid, infoDoc.UID)
	assert.Equal(t, 1, infoDoc.DevType)
	assert.Equal(t, 85, infoDoc.Remain)

	// 验证 battery_details 快照更新
	detailDoc, err := testHelper.Storage.GetBatteryDetail(ctx, uid)
	require.NoError(t, err)
	require.NotNil(t, detailDoc, "battery_details 应有快照")
	assert.Equal(t, "callback", detailDoc.SyncSource)
}

// TestAlarmCallback 验证 alarm 推送正确存入 callback_alarms
func TestAlarmCallback(t *testing.T) {
	uid := fmt.Sprintf("test_uid_%d", time.Now().UnixNano())
	defer cleanupByUID(t, "callback_alarms", uid)

	payload := callback.AlarmPushPayload{
		AppID: "test_app",
		Data: []callback.AlarmPushItem{
			{
				UID:     uid,
				InfoObj: map[string]any{"key": "value"},
				Events: []callback.AlarmEvent{
					{Alarm: "overvoltage", Type: 1, Time: "2026-01-01 12:00:00"},
					{Alarm: "undervoltage", Type: 2, Time: "2026-01-01 12:01:00"},
				},
			},
		},
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/callback/push", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// 验证 callback_alarms 有 2 条记录
	ctx := context.Background()
	col := testHelper.Storage.Collection("callback_alarms")
	count, err := col.CountDocuments(ctx, bson.D{{Key: "uid", Value: uid}})
	require.NoError(t, err)
	assert.Equal(t, int64(2), count, "应有 2 条 alarm 记录")
}

// TestOnlineCallback 验证 online 推送正确存入 callback_online
func TestOnlineCallback(t *testing.T) {
	uid := fmt.Sprintf("test_uid_%d", time.Now().UnixNano())
	defer cleanupByUID(t, "callback_online", uid)

	payload := callback.OnlinePushPayload{
		AppID: "test_app",
		Data: []callback.OnlinePushItem{
			{UID: uid, Online: 1, Time: "2026-01-01 12:00:00"},
		},
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/callback/push", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// 验证 callback_online
	ctx := context.Background()
	col := testHelper.Storage.Collection("callback_online")
	var onlineDoc storage.CallbackOnlineDoc
	err := col.FindOne(ctx, bson.D{{Key: "uid", Value: uid}}).Decode(&onlineDoc)
	require.NoError(t, err)
	assert.Equal(t, uid, onlineDoc.UID)
	assert.Equal(t, 1, onlineDoc.Online)
}

// TestInvalidJSON 验证无效 JSON 返回 400
func TestInvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/callback/push", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestGetRequest 验证 GET 请求返回 405
func TestGetRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/callback/push", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

// TestEmptyData 验证空 data 数组优雅处理
func TestEmptyData(t *testing.T) {
	payload := map[string]any{
		"appid": "test_app",
		"data":  []any{},
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/callback/push", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestUnrecognizedType 验证无法识别类型不报错
func TestUnrecognizedType(t *testing.T) {
	payload := map[string]any{
		"appid": "test_app",
		"data": []map[string]any{
			{"uid": "test_uid", "unknown_field": "value"},
		},
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/callback/push", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ==================== helpers ====================

func cleanupByUID(t *testing.T, collection, uid string) {
	t.Helper()
	ctx := context.Background()
	col := testHelper.Storage.Collection(collection)
	_, err := col.DeleteMany(ctx, bson.D{{Key: "uid", Value: uid}})
	if err != nil {
		t.Logf("清理 %s/%s 失败: %v", collection, uid, err)
	}
}
