package device

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"quanyu-battery-sync/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestProvider_GetUIDs_MockHTTP(t *testing.T) {
	// 模拟正常 API 响应
	uids := []string{"uid001", "uid002", "uid003"}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(uids)
	}))
	defer server.Close()

	cfg := config.DeviceAPIConfig{
		URL:             server.URL,
		RefreshInterval: 1 * time.Hour,
	}

	provider := NewProvider(cfg, zap.NewNop())
	err := provider.Start(context.Background())
	require.NoError(t, err)

	// 等待首次加载完成
	time.Sleep(100 * time.Millisecond)

	result := provider.GetUIDs()
	assert.Equal(t, uids, result)
}

func TestProvider_GetUIDs_EmptyOnAPIError(t *testing.T) {
	// API 返回错误
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cfg := config.DeviceAPIConfig{
		URL:             server.URL,
		RefreshInterval: 1 * time.Hour,
	}

	provider := NewProvider(cfg, zap.NewNop())
	_ = provider.Start(context.Background())

	// 首次加载失败，应返回空列表
	time.Sleep(100 * time.Millisecond)
	result := provider.GetUIDs()
	assert.Empty(t, result)
}

func TestProvider_APIFailure_NoOverwrite(t *testing.T) {
	// 先返回正常数据
	callCount := 0
	uids := []string{"uid001", "uid002"}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			json.NewEncoder(w).Encode(uids)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	cfg := config.DeviceAPIConfig{
		URL:             server.URL,
		RefreshInterval: 100 * time.Millisecond, // 快速刷新
	}

	provider := NewProvider(cfg, zap.NewNop())
	err := provider.Start(context.Background())
	require.NoError(t, err)

	// 等待首次加载完成
	time.Sleep(200 * time.Millisecond)

	// 第一次应该有数据
	result := provider.GetUIDs()
	assert.Equal(t, uids, result)

	// 等待第二次刷新（会失败）
	time.Sleep(200 * time.Millisecond)

	// 失败后应保留旧数据
	result = provider.GetUIDs()
	assert.Equal(t, uids, result, "API 失败后不应覆盖已有数据")
}

func TestProvider_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not valid json"))
	}))
	defer server.Close()

	cfg := config.DeviceAPIConfig{
		URL:             server.URL,
		RefreshInterval: 1 * time.Hour,
	}

	provider := NewProvider(cfg, zap.NewNop())
	err := provider.Start(context.Background())
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	// JSON 解析失败，列表应保持为空
	result := provider.GetUIDs()
	assert.Empty(t, result)
}
