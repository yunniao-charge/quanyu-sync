package sync

import (
	"context"
	"errors"
	"testing"

	"quanyu-battery-sync/internal/config"
	"quanyu-battery-sync/internal/quanyu"
	"quanyu-battery-sync/internal/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newTestSyncer(client *mockQuanyuClient, store *mockSyncStorage, provider *mockDeviceProvider) *Syncer {
	return NewSyncer(client, store, provider, config.SyncConfig{}, zap.NewNop())
}

func TestSyncDetailTask_Success(t *testing.T) {
	var apiCalled, storageCalled bool
	var storedDoc *storage.BatteryDetailDoc

	client := &mockQuanyuClient{
		getBatteryDetailFn: func(ctx context.Context, uid string) (*quanyu.BatteryDetail, error) {
			apiCalled = true
			assert.Equal(t, "UID001", uid)
			return &quanyu.BatteryDetail{
				UID:     "UID001",
				BTCode:  "SN001",
				SOC:     85,
				Voltage: "52.1",
			}, nil
		},
	}

	store := noopStore()
	store.upsertBatteryDetailFn = func(ctx context.Context, doc *storage.BatteryDetailDoc) error {
		storageCalled = true
		storedDoc = doc
		return nil
	}

	s := newTestSyncer(client, store, &mockDeviceProvider{})
	err := s.syncDetailTask(context.Background(), "UID001")

	require.NoError(t, err)
	assert.True(t, apiCalled, "API should be called")
	assert.True(t, storageCalled, "Storage should be called")
	assert.Equal(t, "UID001", storedDoc.UID)
	assert.Equal(t, "SN001", storedDoc.SN)
	assert.Equal(t, 85, storedDoc.SOC)
	assert.Equal(t, "52.1", storedDoc.Voltage)
	assert.Equal(t, "pull", storedDoc.SyncSource)
}

func TestSyncDetailTask_APIError(t *testing.T) {
	var errorUpdated bool
	var capturedErrMsg string

	client := &mockQuanyuClient{
		getBatteryDetailFn: func(ctx context.Context, uid string) (*quanyu.BatteryDetail, error) {
			return nil, errors.New("API timeout")
		},
	}

	store := noopStore()
	store.updateSyncErrorFn = func(ctx context.Context, uid, syncType, errMsg string) error {
		errorUpdated = true
		assert.Equal(t, "UID001", uid)
		assert.Equal(t, "detail", syncType)
		capturedErrMsg = errMsg
		return nil
	}

	s := newTestSyncer(client, store, &mockDeviceProvider{})
	err := s.syncDetailTask(context.Background(), "UID001")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "API timeout")
	assert.True(t, errorUpdated, "UpdateSyncError should be called")
	assert.Contains(t, capturedErrMsg, "API timeout")
}
