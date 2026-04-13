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

func TestSyncChargeTask_MultiPage(t *testing.T) {
	callPages := []int{}
	client := &mockQuanyuClient{
		getChargeRecordsFn: func(ctx context.Context, uid, beginStart, beginEnd string, page, limit int) (*quanyu.ChargeDataResponse, error) {
			callPages = append(callPages, page)
			switch page {
			case 1:
				return &quanyu.ChargeDataResponse{
					TotalPage: 2,
					List:      make([]quanyu.ChargeRecordItem, 100),
				}, nil
			case 2:
				return &quanyu.ChargeDataResponse{
					TotalPage: 2,
					List:      make([]quanyu.ChargeRecordItem, 30),
				}, nil
			default:
				t.Fatalf("unexpected page: %d", page)
				return nil, nil
			}
		},
	}

	var storedCount int
	store := noopStore()
	store.upsertChargeRecordsFn = func(ctx context.Context, docs []storage.ChargeRecordDoc) (int, error) {
		storedCount = len(docs)
		return len(docs), nil
	}

	s := newTestSyncer(client, store, &mockDeviceProvider{})
	err := s.syncChargeDataTask(context.Background(), "UID001")

	require.NoError(t, err)
	assert.Equal(t, []int{1, 2}, callPages, "should paginate through all pages")
	assert.Equal(t, 130, storedCount, "should collect all charge records")
}

func TestSyncChargeTask_DefaultRange(t *testing.T) {
	var capturedStart string

	client := &mockQuanyuClient{
		getChargeRecordsFn: func(ctx context.Context, uid, beginStart, beginEnd string, page, limit int) (*quanyu.ChargeDataResponse, error) {
			if page == 1 {
				capturedStart = beginStart
			}
			return &quanyu.ChargeDataResponse{List: nil}, nil
		},
	}

	store := noopStore()
	s := newTestSyncer(client, store, &mockDeviceProvider{})
	err := s.syncChargeDataTask(context.Background(), "UID001")

	require.NoError(t, err)
	startTime, err := time.ParseInLocation("2006-01-02 15:04:05", capturedStart, time.Local)
	require.NoError(t, err)
	diff := time.Since(startTime)
	assert.True(t, diff >= 2*time.Hour-5*time.Second && diff <= 2*time.Hour+5*time.Second,
		"startTime should be ~2 hours before now, got diff=%v", diff)
}

func TestSyncChargeTask_IncrementalFromState(t *testing.T) {
	var capturedStart string
	lastSyncTime := "2026-04-13 08:00:00"

	client := &mockQuanyuClient{
		getChargeRecordsFn: func(ctx context.Context, uid, beginStart, beginEnd string, page, limit int) (*quanyu.ChargeDataResponse, error) {
			if page == 1 {
				capturedStart = beginStart
			}
			return &quanyu.ChargeDataResponse{List: nil}, nil
		},
	}

	store := noopStore()
	store.getSyncStateFn = func(ctx context.Context, uid, syncType string) (*storage.SyncStateDoc, error) {
		return &storage.SyncStateDoc{LastSyncTime: lastSyncTime}, nil
	}

	s := newTestSyncer(client, store, &mockDeviceProvider{})
	err := s.syncChargeDataTask(context.Background(), "UID001")

	require.NoError(t, err)
	assert.Equal(t, lastSyncTime, capturedStart)
}

func TestSyncChargeTask_CustomDefaultRange(t *testing.T) {
	var capturedStart string

	client := &mockQuanyuClient{
		getChargeRecordsFn: func(ctx context.Context, uid, beginStart, beginEnd string, page, limit int) (*quanyu.ChargeDataResponse, error) {
			if page == 1 {
				capturedStart = beginStart
			}
			return &quanyu.ChargeDataResponse{List: nil}, nil
		},
	}

	store := noopStore()

	cfg := config.SyncConfig{
		ChargeData: config.SyncTaskConfig{DefaultRange: 4 * time.Hour},
	}
	s := NewSyncer(client, store, &mockDeviceProvider{}, cfg, zap.NewNop())
	err := s.syncChargeDataTask(context.Background(), "UID001")

	require.NoError(t, err)
	startTime, err := time.ParseInLocation("2006-01-02 15:04:05", capturedStart, time.Local)
	require.NoError(t, err)
	diff := time.Since(startTime)
	assert.True(t, diff >= 4*time.Hour-5*time.Second && diff <= 4*time.Hour+5*time.Second,
		"startTime should be ~4 hours before now, got diff=%v", diff)
}

func TestSyncChargeTask_APIError(t *testing.T) {
	client := &mockQuanyuClient{
		getChargeRecordsFn: func(ctx context.Context, uid, beginStart, beginEnd string, page, limit int) (*quanyu.ChargeDataResponse, error) {
			return nil, errors.New("server error")
		},
	}

	var errorUpdated bool
	store := noopStore()
	store.updateSyncErrorFn = func(ctx context.Context, uid, syncType, errMsg string) error {
		errorUpdated = true
		assert.Equal(t, "charge_data", syncType)
		return nil
	}

	s := newTestSyncer(client, store, &mockDeviceProvider{})
	err := s.syncChargeDataTask(context.Background(), "UID001")

	require.Error(t, err)
	assert.True(t, errorUpdated)
}
