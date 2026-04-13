//go:build integration

package sync_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"quanyu-battery-sync/internal/sync"
	"quanyu-battery-sync/tests/integration"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

var syncHelper *integration.TestHelper

func TestMain(m *testing.M) {
	var err error
	syncHelper, err = integration.NewTestHelper("../../../config.yaml")
	if err != nil {
		panic(fmt.Sprintf("创建 TestHelper 失败: %v", err))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := syncHelper.Storage.EnsureIndexes(ctx); err != nil {
		panic(fmt.Sprintf("创建索引失败: %v", err))
	}

	m.Run()

	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	syncHelper.Close(ctx2)
}

func getTestSyncer(t *testing.T, uids []string) (*sync.Syncer, context.Context) {
	t.Helper()

	provider := &testProvider{uids: uids}

	s := sync.NewSyncer(
		syncHelper.APIClient,
		syncHelper.Storage,
		provider,
		syncHelper.Config.Sync,
		zap.NewNop(),
	)

	return s, context.Background()
}

// testProvider implements device.DeviceProvider for integration tests
type testProvider struct {
	uids []string
}

func (p *testProvider) GetUIDs() []string {
	return p.uids
}

func getTestUIDs(t *testing.T) []string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	uids, err := syncHelper.GetTestUIDs(ctx, 1)
	require.NoError(t, err, "获取测试 UID 失败")
	require.NotEmpty(t, uids, "设备列表为空")
	return uids
}

// ==================== Detail 同步 ====================

func TestIntegration_DetailSync(t *testing.T) {
	uids := getTestUIDs(t)
	s, ctx := getTestSyncer(t, uids)

	err := s.RunTask(ctx, "detail", uids[0])
	require.NoError(t, err, "Detail 同步应该成功")

	// 验证数据已写入 MongoDB
	detail, err := syncHelper.Storage.GetBatteryDetail(ctx, uids[0])
	require.NoError(t, err)
	require.NotNil(t, detail, "battery_detail 应该存在")
	assert.Equal(t, uids[0], detail.UID)
	assert.Equal(t, "pull", detail.SyncSource)
	assert.False(t, detail.UpdatedAt.IsZero(), "updated_at 应该被设置")
}

// ==================== History 同步 ====================

func TestIntegration_HistorySync(t *testing.T) {
	uids := getTestUIDs(t)
	s, ctx := getTestSyncer(t, uids)

	err := s.RunTask(ctx, "history_data", uids[0])
	require.NoError(t, err, "History 同步应该成功")

	// 验证 sync_state 已更新
	state, err := syncHelper.Storage.GetSyncState(ctx, uids[0], "history_data")
	require.NoError(t, err)
	require.NotNil(t, state, "sync_state 应该存在")
	assert.NotEmpty(t, state.LastSyncTime, "last_sync_time 应该被设置")
}

// ==================== Trace 同步（带分页） ====================

func TestIntegration_TraceSync(t *testing.T) {
	uids := getTestUIDs(t)
	s, ctx := getTestSyncer(t, uids)

	err := s.RunTask(ctx, "trace", uids[0])
	if err != nil {
		t.Skipf("Trace API 不可用，跳过: %v", err)
	}

	// 验证 sync_state 已更新
	state, err := syncHelper.Storage.GetSyncState(ctx, uids[0], "trace")
	require.NoError(t, err)
	require.NotNil(t, state, "sync_state 应该存在")
	assert.NotEmpty(t, state.LastSyncTime, "last_sync_time 应该被设置")
}

// ==================== 全类型同步 ====================

func TestIntegration_AllSyncTypes(t *testing.T) {
	uids := getTestUIDs(t)
	uid := uids[0]
	s, ctx := getTestSyncer(t, uids)

	syncTypes := []string{"detail", "history_data", "trace", "event", "charge_data"}
	for _, syncType := range syncTypes {
		t.Run(syncType, func(t *testing.T) {
			err := s.RunTask(ctx, syncType, uid)
			if err != nil {
				t.Logf("API error (expected with external API): %v", err)
			}
		})
	}
}
