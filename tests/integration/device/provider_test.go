//go:build integration

package device_test

import (
	"context"
	"fmt"
	"testing"
	"time"

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

	m.Run()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	testHelper.Close(ctx)
}

func TestProvider_RealAPI_GetUIDs(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	uids, err := testHelper.GetTestUIDs(ctx, 10)
	require.NoError(t, err, "获取 UID 列表失败")

	assert.NotEmpty(t, uids, "UID 列表不应为空")
	t.Logf("获取到 %d 个 UID: %v", len(uids), uids)

	// 验证 UID 格式
	for _, uid := range uids {
		assert.NotEmpty(t, uid, "UID 不应为空字符串")
	}
}
