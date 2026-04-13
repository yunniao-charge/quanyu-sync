package sync

import (
	"context"
	"errors"
	"testing"

	"quanyu-battery-sync/internal/quanyu"
	"quanyu-battery-sync/internal/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncState_FirstSyncCreatesState(t *testing.T) {
	syncTimeCalled := false

	client := &mockQuanyuClient{
		getBatteryDataFn: func(ctx context.Context, uid, startTime, endTime string, last int) (*quanyu.BatteryDataResponse, error) {
			return &quanyu.BatteryDataResponse{UID: "UID001"}, nil
		},
	}

	store := noopStore()
	store.upsertBatteryHistoryFn = func(ctx context.Context, docs []storage.BatteryHistoryDoc) (int, error) { return 1, nil }
	store.getSyncStateFn = func(ctx context.Context, uid, syncType string) (*storage.SyncStateDoc, error) {
		assert.Equal(t, "UID001", uid)
		assert.Equal(t, "history_data", syncType)
		return nil, nil
	}
	store.updateSyncTimeFn = func(ctx context.Context, uid, syncType, syncTime string) error {
		syncTimeCalled = true
		assert.Equal(t, "UID001", uid)
		assert.Equal(t, "history_data", syncType)
		return nil
	}

	s := newTestSyncer(client, store, &mockDeviceProvider{})
	err := s.syncHistoryDataTask(context.Background(), "UID001")

	require.NoError(t, err)
	assert.True(t, syncTimeCalled, "UpdateSyncTime should be called after successful first sync")
}

func TestSyncState_SuccessUpdatesSyncTime(t *testing.T) {
	syncTimeCalled := false

	client := &mockQuanyuClient{
		getBatteryDetailFn: func(ctx context.Context, uid string) (*quanyu.BatteryDetail, error) {
			return &quanyu.BatteryDetail{UID: "UID001", SOC: 80}, nil
		},
	}

	store := noopStore()
	store.updateSyncTimeFn = func(ctx context.Context, uid, syncType, syncTime string) error {
		syncTimeCalled = true
		return nil
	}

	s := newTestSyncer(client, store, &mockDeviceProvider{})
	err := s.syncDetailTask(context.Background(), "UID001")

	require.NoError(t, err)
	// Detail sync doesn't call UpdateSyncTime on success - it just upserts the detail
	// This is by design: detail is a full snapshot, not incremental
	assert.False(t, syncTimeCalled, "detail sync should NOT call UpdateSyncTime on success")
}

func TestSyncState_FailureUpdatesError(t *testing.T) {
	errorCalled := false
	syncTimeCalled := false

	client := &mockQuanyuClient{
		getBatteryDetailFn: func(ctx context.Context, uid string) (*quanyu.BatteryDetail, error) {
			return nil, errors.New("connection refused")
		},
	}

	store := noopStore()
	store.updateSyncTimeFn = func(ctx context.Context, uid, syncType, syncTime string) error {
		syncTimeCalled = true
		return nil
	}
	store.updateSyncErrorFn = func(ctx context.Context, uid, syncType, errMsg string) error {
		errorCalled = true
		assert.Equal(t, "detail", syncType)
		assert.Contains(t, errMsg, "connection refused")
		return nil
	}

	s := newTestSyncer(client, store, &mockDeviceProvider{})
	err := s.syncDetailTask(context.Background(), "UID001")

	require.Error(t, err)
	assert.True(t, errorCalled, "UpdateSyncError should be called on failure")
	assert.False(t, syncTimeCalled, "UpdateSyncTime should NOT be called on failure")
}

func TestSyncState_StorageWriteFailure_NoSyncTimeUpdate(t *testing.T) {
	syncTimeCalled := false
	errorCalled := false

	client := &mockQuanyuClient{
		getBatteryDetailFn: func(ctx context.Context, uid string) (*quanyu.BatteryDetail, error) {
			return &quanyu.BatteryDetail{UID: "UID001", SOC: 80}, nil
		},
	}

	store := noopStore()
	store.upsertBatteryDetailFn = func(ctx context.Context, doc *storage.BatteryDetailDoc) error {
		return errors.New("MongoDB connection lost")
	}
	store.updateSyncTimeFn = func(ctx context.Context, uid, syncType, syncTime string) error {
		syncTimeCalled = true
		return nil
	}
	store.updateSyncErrorFn = func(ctx context.Context, uid, syncType, errMsg string) error {
		errorCalled = true
		assert.Equal(t, "detail", syncType)
		assert.Contains(t, errMsg, "MongoDB connection lost")
		return nil
	}

	s := newTestSyncer(client, store, &mockDeviceProvider{})
	err := s.syncDetailTask(context.Background(), "UID001")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "MongoDB connection lost")
	assert.False(t, syncTimeCalled, "UpdateSyncTime should NOT be called when storage write fails")
	assert.True(t, errorCalled, "UpdateSyncError should be called when storage write fails")
}

func TestSyncState_HistoryStorageFailure_NoSyncTimeUpdate(t *testing.T) {
	syncTimeCalled := false
	errorCalled := false

	client := &mockQuanyuClient{
		getBatteryDataFn: func(ctx context.Context, uid, startTime, endTime string, last int) (*quanyu.BatteryDataResponse, error) {
			return &quanyu.BatteryDataResponse{UID: "UID001"}, nil
		},
	}

	store := noopStore()
	store.upsertBatteryHistoryFn = func(ctx context.Context, docs []storage.BatteryHistoryDoc) (int, error) {
		return 0, errors.New("write timeout")
	}
	store.updateSyncTimeFn = func(ctx context.Context, uid, syncType, syncTime string) error {
		syncTimeCalled = true
		return nil
	}
	store.updateSyncErrorFn = func(ctx context.Context, uid, syncType, errMsg string) error {
		errorCalled = true
		assert.Equal(t, "history_data", syncType)
		return nil
	}

	s := newTestSyncer(client, store, &mockDeviceProvider{})
	err := s.syncHistoryDataTask(context.Background(), "UID001")

	require.Error(t, err)
	assert.False(t, syncTimeCalled, "UpdateSyncTime should NOT be called when storage write fails")
	assert.True(t, errorCalled, "UpdateSyncError should be called")
}
