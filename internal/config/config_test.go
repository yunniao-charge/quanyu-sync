package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_NormalConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	content := `
quanyu:
  appid: "test_appid"
  key: "test_key"
  base_url: "https://iot.xtrunc.com"
  timeout: 30s
  max_retries: 3

device_api:
  url: "https://api.example.com/uids"
  refresh_interval: 6h

mongodb:
  uri: "mongodb://localhost:27017"
  database: "test_db"
  username: "root"
  password: "root123"

callback:
  listen_addr: ":8888"
  notify_url: "http://example.com:8888/callback/push"

subscribe:
  batch_size: 20
  renew_interval: 25m
  sub_data:
    - "info"
    - "alarm"

log:
  dir: "logs"
  level: "debug"
  max_size: 50
  max_backups: 10
  max_age: 7

sync:
  detail:
    cron: "0 */2 * * * *"
    enabled: true
  history_data:
    cron: "0 */5 * * * *"
    enabled: true
    default_range: 1h
`
	require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0644))

	cfg, err := Load(cfgPath)
	require.NoError(t, err)

	// 验证所有字段正确映射
	assert.Equal(t, "test_appid", cfg.Quanyu.AppID)
	assert.Equal(t, "test_key", cfg.Quanyu.Key)
	assert.Equal(t, "https://iot.xtrunc.com", cfg.Quanyu.BaseURL)
	assert.Equal(t, 30*time.Second, cfg.Quanyu.Timeout)
	assert.Equal(t, 3, cfg.Quanyu.MaxRetries)

	assert.Equal(t, "https://api.example.com/uids", cfg.DeviceAPI.URL)
	assert.Equal(t, 6*time.Hour, cfg.DeviceAPI.RefreshInterval)

	assert.Equal(t, "mongodb://localhost:27017", cfg.MongoDB.URI)
	assert.Equal(t, "test_db", cfg.MongoDB.Database)
	assert.Equal(t, "root", cfg.MongoDB.Username)
	assert.Equal(t, "root123", cfg.MongoDB.Password)

	assert.Equal(t, ":8888", cfg.Callback.ListenAddr)
	assert.Equal(t, "http://example.com:8888/callback/push", cfg.Callback.NotifyURL)

	assert.Equal(t, 20, cfg.Subscribe.BatchSize)
	assert.Equal(t, 25*time.Minute, cfg.Subscribe.RenewInterval)
	assert.Equal(t, []string{"info", "alarm"}, cfg.Subscribe.SubData)

	assert.Equal(t, "logs", cfg.Log.Dir)
	assert.Equal(t, "debug", cfg.Log.Level)
	assert.Equal(t, 50, cfg.Log.MaxSize)
	assert.Equal(t, 10, cfg.Log.MaxBackups)
	assert.Equal(t, 7, cfg.Log.MaxAge)

	assert.Equal(t, "0 */2 * * * *", cfg.Sync.Detail.Cron)
	assert.True(t, cfg.Sync.Detail.Enabled)
	assert.Equal(t, "0 */5 * * * *", cfg.Sync.HistoryData.Cron)
	assert.Equal(t, 1*time.Hour, cfg.Sync.HistoryData.DefaultRange)
}

func TestLoad_DefaultValues(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	// 最小配置 - 所有可选字段缺失
	content := `
quanyu:
  appid: "test_appid"
  key: "test_key"
`
	require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0644))

	cfg, err := Load(cfgPath)
	require.NoError(t, err)

	// 验证默认值
	assert.Equal(t, "https://iot.xtrunc.com", cfg.Quanyu.BaseURL)
	assert.Equal(t, 30*time.Second, cfg.Quanyu.Timeout)
	assert.Equal(t, 3, cfg.Quanyu.MaxRetries)

	assert.Equal(t, 6*time.Hour, cfg.DeviceAPI.RefreshInterval)

	assert.Equal(t, "mongodb://localhost:27017", cfg.MongoDB.URI)
	assert.Equal(t, "quanyu_battery", cfg.MongoDB.Database)

	assert.Equal(t, ":8888", cfg.Callback.ListenAddr)

	assert.Equal(t, 20, cfg.Subscribe.BatchSize)
	assert.Equal(t, 25*time.Minute, cfg.Subscribe.RenewInterval)
	assert.Equal(t, []string{"info", "alarm", "online"}, cfg.Subscribe.SubData)

	assert.Equal(t, "logs", cfg.Log.Dir)
	assert.Equal(t, "info", cfg.Log.Level)
	assert.Equal(t, 100, cfg.Log.MaxSize)
	assert.Equal(t, 30, cfg.Log.MaxBackups)
	assert.Equal(t, 30, cfg.Log.MaxAge)

	assert.Equal(t, "0 */2 * * * *", cfg.Sync.Detail.Cron)
	assert.Equal(t, "0 */5 * * * *", cfg.Sync.HistoryData.Cron)
	assert.Equal(t, "0 */5 * * * *", cfg.Sync.Trace.Cron)
	assert.Equal(t, "0 */5 * * * *", cfg.Sync.Event.Cron)
	assert.Equal(t, "0 */10 * * * *", cfg.Sync.ChargeData.Cron)

	assert.Equal(t, 1*time.Hour, cfg.Sync.HistoryData.DefaultRange)
	assert.Equal(t, 1*time.Hour, cfg.Sync.Trace.DefaultRange)
	assert.Equal(t, 1*time.Hour, cfg.Sync.Event.DefaultRange)
	assert.Equal(t, 2*time.Hour, cfg.Sync.ChargeData.DefaultRange)
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	content := `
quanyu:
  appid: [invalid yaml
  key: missing bracket
`
	require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0644))

	_, err := Load(cfgPath)
	assert.Error(t, err, "无效 YAML 应返回错误")
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	assert.Error(t, err, "不存在的文件应返回错误")
}
