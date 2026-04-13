package sync

import (
	"context"
	"errors"
	"testing"
	"time"

	"quanyu-battery-sync/internal/config"
	"quanyu-battery-sync/internal/quanyu"
	"quanyu-battery-sync/internal/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestSyncHistoryTask_Success(t *testing.T) {
	var capturedStart, capturedEnd string
	var storedDocs []storage.BatteryHistoryDoc
	var syncTimeUpdated bool
	var capturedSyncTime string

	client := &mockQuanyuClient{
		getBatteryDataFn: func(ctx context.Context, uid, startTime, endTime string, last int) (*quanyu.BatteryDataResponse, error) {
			capturedStart = startTime
			capturedEnd = endTime
			assert.Equal(t, "UID001", uid)
			return &quanyu.BatteryDataResponse{
				UID:       "UID001",
				TimeScale: "scale1",
				VVV:       "52.1",
				Cap:       "100",
				Rem:       "85",
			}, nil
		},
	}

	store := noopStore()
	store.upsertBatteryHistoryFn = func(ctx context.Context, docs []storage.BatteryHistoryDoc) (int, error) {
		storedDocs = docs
		return 1, nil
	}
	store.updateSyncTimeFn = func(ctx context.Context, uid, syncType, syncTime string) error {
		syncTimeUpdated = true
		capturedSyncTime = syncTime
		return nil
	}

	s := newTestSyncer(client, store, &mockDeviceProvider{})
	err := s.syncHistoryDataTask(context.Background(), "UID001")

	require.NoError(t, err)

	// Default time range should be ~1 hour back
	_, err = time.ParseInLocation("2006-01-02 15:04:05", capturedStart, time.Local)
	require.NoError(t, err, "startTime should be valid datetime format")
	_, err = time.ParseInLocation("2006-01-02 15:04:05", capturedEnd, time.Local)
	require.NoError(t, err, "endTime should be valid datetime format")

	// Data stored correctly
	require.Len(t, storedDocs, 1)
	assert.Equal(t, "UID001", storedDocs[0].UID)
	assert.Equal(t, "52.1", storedDocs[0].VVV)
	assert.Equal(t, "100", storedDocs[0].Cap)

	// Sync time updated
	assert.True(t, syncTimeUpdated)
	_, err = time.ParseInLocation("2006-01-02 15:04:05", capturedSyncTime, time.Local)
	assert.NoError(t, err, "syncTime should be valid datetime format")
}

func TestSyncHistoryTask_IncrementalFromLastSync(t *testing.T) {
	var capturedStart string
	lastSyncTime := "2026-04-13 10:00:00"

	client := &mockQuanyuClient{
		getBatteryDataFn: func(ctx context.Context, uid, startTime, endTime string, last int) (*quanyu.BatteryDataResponse, error) {
			capturedStart = startTime
			return &quanyu.BatteryDataResponse{UID: "UID001"}, nil
		},
	}

	store := noopStore()
	store.upsertBatteryHistoryFn = func(ctx context.Context, docs []storage.BatteryHistoryDoc) (int, error) { return 1, nil }
	store.getSyncStateFn = func(ctx context.Context, uid, syncType string) (*storage.SyncStateDoc, error) {
		return &storage.SyncStateDoc{
			UID:          "UID001",
			SyncType:     "history_data",
			LastSyncTime: lastSyncTime,
		}, nil
	}

	s := newTestSyncer(client, store, &mockDeviceProvider{})
	err := s.syncHistoryDataTask(context.Background(), "UID001")

	require.NoError(t, err)
	assert.Equal(t, lastSyncTime, capturedStart, "should use last_sync_time as start")
}

func TestSyncHistoryTask_DefaultRange(t *testing.T) {
	var capturedStart string

	client := &mockQuanyuClient{
		getBatteryDataFn: func(ctx context.Context, uid, startTime, endTime string, last int) (*quanyu.BatteryDataResponse, error) {
			capturedStart = startTime
			return &quanyu.BatteryDataResponse{UID: "UID001"}, nil
		},
	}

	store := noopStore()
	store.upsertBatteryHistoryFn = func(ctx context.Context, docs []storage.BatteryHistoryDoc) (int, error) { return 1, nil }

	cfg := config.SyncConfig{
		HistoryData: config.SyncTaskConfig{
			DefaultRange: 2 * time.Hour,
		},
	}
	s := NewSyncer(client, store, &mockDeviceProvider{}, cfg, zap.NewNop())
	err := s.syncHistoryDataTask(context.Background(), "UID001")

	require.NoError(t, err)

	startTime, err := time.ParseInLocation("2006-01-02 15:04:05", capturedStart, time.Local)
	require.NoError(t, err)
	now := time.Now()
	diff := now.Sub(startTime)
	assert.True(t, diff >= 2*time.Hour-5*time.Second && diff <= 2*time.Hour+5*time.Second,
		"startTime should be ~2 hours before now, got diff=%v", diff)
}

func TestSyncHistoryTask_APIError(t *testing.T) {
	client := &mockQuanyuClient{
		getBatteryDataFn: func(ctx context.Context, uid, startTime, endTime string, last int) (*quanyu.BatteryDataResponse, error) {
			return nil, errors.New("network error")
		},
	}

	var errorUpdated bool
	store := noopStore()
	store.updateSyncErrorFn = func(ctx context.Context, uid, syncType, errMsg string) error {
		errorUpdated = true
		assert.Equal(t, "history_data", syncType)
		return nil
	}

	s := newTestSyncer(client, store, &mockDeviceProvider{})
	err := s.syncHistoryDataTask(context.Background(), "UID001")

	require.Error(t, err)
	assert.True(t, errorUpdated, "UpdateSyncError should be called on API error")
}
