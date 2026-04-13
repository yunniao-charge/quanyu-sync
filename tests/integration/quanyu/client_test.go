//go:build integration

package quanyu_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	"quanyu-battery-sync/internal/quanyu"
	"quanyu-battery-sync/tests/integration"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testHelper *integration.TestHelper

func TestMain(m *testing.M) {
	var err error
	testHelper, err = integration.NewTestHelper("../../../config.yaml")
	if err != nil {
		panic(fmt.Sprintf("创建 TestHelper 失败: %v", err))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 确保索引存在
	if err := testHelper.Storage.EnsureIndexes(ctx); err != nil {
		panic(fmt.Sprintf("创建索引失败: %v", err))
	}

	m.Run()

	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	testHelper.Close(ctx2)
}

func getTestUID(t *testing.T) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	uids, err := testHelper.GetTestUIDs(ctx, 1)
	require.NoError(t, err, "获取测试 UID 失败")
	require.NotEmpty(t, uids, "设备列表为空")
	return uids[0]
}

// TestGetBatteryDetail 验证 device/detail 接口
func TestGetBatteryDetail(t *testing.T) {
	uid := getTestUID(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	detail, err := testHelper.APIClient.GetBatteryDetail(ctx, uid)
	require.NoError(t, err, "GetBatteryDetail 调用失败")

	assert.NotEmpty(t, detail.UID, "UID 不应为空")
	assert.NotEmpty(t, detail.Voltage, "Voltage 不应为空")
	t.Logf("BatteryDetail: uid=%s, soc=%d, voltage=%s", detail.UID, detail.SOC, detail.Voltage)
}

// TestGetBatteryData 验证 device/data 接口（时间范围参数）
func TestGetBatteryData(t *testing.T) {
	uid := getTestUID(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	startTime := time.Now().Add(-1 * time.Hour).Format("2006-01-02 15:04:05")
	endTime := time.Now().Format("2006-01-02 15:04:05")

	data, err := testHelper.APIClient.GetBatteryData(ctx, uid, startTime, endTime, 0)
	require.NoError(t, err, "GetBatteryData 调用失败")

	t.Logf("BatteryData: uid=%s", data.UID)
}

// TestGetBatteryTrace 验证 battrace 接口（分页参数）
func TestGetBatteryTrace(t *testing.T) {
	uid := getTestUID(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	startTime := time.Now().Add(-1 * time.Hour).Format("20060102150405")
	endTime := time.Now().Format("20060102150405")

	trace, err := testHelper.APIClient.GetBatteryTrace(ctx, uid, startTime, endTime, 1, 50)
	require.NoError(t, err, "GetBatteryTrace 调用失败")

	t.Logf("BatteryTrace: uid=%s, pages=%d, trace_points=%d", trace.UID, trace.Pages, len(trace.Trace))
}

// TestGetBatteryEvents 验证 device/eventData 接口
func TestGetBatteryEvents(t *testing.T) {
	uid := getTestUID(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	startTime := time.Now().Add(-1 * time.Hour).Format("2006-01-02 15:04:05")
	endTime := time.Now().Format("2006-01-02 15:04:05")

	events, err := testHelper.APIClient.GetBatteryEvents(ctx, uid, startTime, endTime, 1, 50)
	require.NoError(t, err, "GetBatteryEvents 调用失败")

	t.Logf("BatteryEvents: total=%d, list_count=%d", events.Total, len(events.List))
}

// TestGetChargeRecords 验证 device/chargeData 接口
func TestGetChargeRecords(t *testing.T) {
	uid := getTestUID(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	startTime := time.Now().Add(-2 * time.Hour).Format("2006-01-02 15:04:05")
	endTime := time.Now().Format("2006-01-02 15:04:05")

	records, err := testHelper.APIClient.GetChargeRecords(ctx, uid, startTime, endTime, 1, 50)
	require.NoError(t, err, "GetChargeRecords 调用失败")

	t.Logf("ChargeRecords: page=%d/%d, total=%d, list_count=%d",
		records.CurrPage, records.TotalPage, records.TotalCount, len(records.List))
}

// TestInvalidUIDErrorHandling 验证无效 UID 的错误处理
func TestInvalidUIDErrorHandling(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	invalidUID := "INVALID_UID_NOT_EXISTS_12345"

	_, err := testHelper.APIClient.GetBatteryDetail(ctx, invalidUID)
	// 无效 UID 应返回错误（errno != 0）
	assert.Error(t, err, "无效 UID 应返回错误")
	assert.Contains(t, err.Error(), "errno", "错误信息应包含 errno")

	t.Logf("Invalid UID error: %v", err)
}

// TestSignFormat 验证 API 请求中的签名字段格式
func TestSignFormat(t *testing.T) {
	appid := "test_appid"
	nonceStr := quanyu.GenerateNonceStr(32)
	uid := "test_uid"
	key := "test_key"

	sign := quanyu.GenerateSign(appid, nonceStr, uid, key)

	// 签名应为 32 位大写 MD5
	matched, err := regexp.MatchString(`^[A-F0-9]{32}$`, sign)
	require.NoError(t, err)
	assert.True(t, matched, "签名应为 32 位大写十六进制: %s", sign)
}

// TestSignDeterministic 验证签名确定性
func TestSignDeterministic(t *testing.T) {
	sign1 := quanyu.GenerateSign("app1", "nonce1", "uid1", "key1")
	sign2 := quanyu.GenerateSign("app1", "nonce1", "uid1", "key1")
	assert.Equal(t, sign1, sign2, "相同输入应产生相同签名")
}
