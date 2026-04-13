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

func TestSyncEventTask_MultiPage(t *testing.T) {
	callPages := []int{}
	client := &mockQuanyuClient{
		getBatteryEventsFn: func(ctx context.Context, uid, startTime, endTime string, page, limit int) (*quanyu.BatteryEventResponse, error) {
			callPages = append(callPages, page)
			switch page {
			case 1:
				return &quanyu.BatteryEventResponse{
					Total: 250,
					List:  make([]quanyu.BatteryEventItem, 100),
				}, nil
			case 2:
				return &quanyu.BatteryEventResponse{
					Total: 250,
					List:  make([]quanyu.BatteryEventItem, 100),
				}, nil
			case 3:
				return &quanyu.BatteryEventResponse{
					Total: 250,
					List:  make([]quanyu.BatteryEventItem, 50),
				}, nil
			default:
				t.Fatalf("unexpected page: %d", page)
				return nil, nil
			}
		},
	}

	var storedCount int
	store := noopStore()
	store.upsertBatteryEventsFn = func(ctx context.Context, docs []storage.BatteryEventDoc) (int, error) {
		storedCount = len(docs)
		return len(docs), nil
	}

	s := newTestSyncer(client, store, &mockDeviceProvider{})
	err := s.syncEventTask(context.Background(), "UID001")

	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, callPages, "should paginate through all pages")
	assert.Equal(t, 250, storedCount, "should collect all events")
}

func TestSyncEventTask_IncrementalTimeRange(t *testing.T) {
	var capturedStart string
	lastSyncTime := "2026-04-13 10:00:00"

	client := &mockQuanyuClient{
		getBatteryEventsFn: func(ctx context.Context, uid, startTime, endTime string, page, limit int) (*quanyu.BatteryEventResponse, error) {
			if page == 1 {
				capturedStart = startTime
			}
			return &quanyu.BatteryEventResponse{List: nil}, nil
		},
	}

	store := noopStore()
	store.getSyncStateFn = func(ctx context.Context, uid, syncType string) (*storage.SyncStateDoc, error) {
		return &storage.SyncStateDoc{LastSyncTime: lastSyncTime}, nil
	}

	s := newTestSyncer(client, store, &mockDeviceProvider{})
	err := s.syncEventTask(context.Background(), "UID001")

	require.NoError(t, err)
	assert.Equal(t, lastSyncTime, capturedStart, "should use last_sync_time as start")
}

func TestSyncEventTask_DefaultRange(t *testing.T) {
	var capturedStart string

	client := &mockQuanyuClient{
		getBatteryEventsFn: func(ctx context.Context, uid, startTime, endTime string, page, limit int) (*quanyu.BatteryEventResponse, error) {
			if page == 1 {
				capturedStart = startTime
			}
			return &quanyu.BatteryEventResponse{List: nil}, nil
		},
	}

	store := noopStore()

	cfg := config.SyncConfig{
		Event: config.SyncTaskConfig{DefaultRange: 3 * time.Hour},
	}
	s := NewSyncer(client, store, &mockDeviceProvider{}, cfg, zap.NewNop())
	err := s.syncEventTask(context.Background(), "UID001")

	require.NoError(t, err)
	startTime, err := time.ParseInLocation("2006-01-02 15:04:05", capturedStart, time.Local)
	require.NoError(t, err)
	diff := time.Since(startTime)
	assert.True(t, diff >= 3*time.Hour-5*time.Second && diff <= 3*time.Hour+5*time.Second,
		"startTime should be ~3 hours before now, got diff=%v", diff)
}

func TestSyncEventTask_APIError(t *testing.T) {
	client := &mockQuanyuClient{
		getBatteryEventsFn: func(ctx context.Context, uid, startTime, endTime string, page, limit int) (*quanyu.BatteryEventResponse, error) {
			return nil, errors.New("network error")
		},
	}

	var errorUpdated bool
	store := noopStore()
	store.updateSyncErrorFn = func(ctx context.Context, uid, syncType, errMsg string) error {
		errorUpdated = true
		assert.Equal(t, "event", syncType)
		return nil
	}

	s := newTestSyncer(client, store, &mockDeviceProvider{})
	err := s.syncEventTask(context.Background(), "UID001")

	require.Error(t, err)
	assert.True(t, errorUpdated)
}
