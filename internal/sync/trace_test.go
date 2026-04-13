package sync

import (
	"context"
	"errors"
	"testing"
	"time"

	"quanyu-battery-sync/internal/quanyu"
	"quanyu-battery-sync/internal/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncTraceTask_MultiPage(t *testing.T) {
	callPages := []int{}
	client := &mockQuanyuClient{
		getBatteryTraceFn: func(ctx context.Context, uid, startTime, endTime string, pageNum, pageSize int) (*quanyu.BatteryTraceResponse, error) {
			callPages = append(callPages, pageNum)
			switch pageNum {
			case 1:
				return &quanyu.BatteryTraceResponse{
					UID:   "UID001",
					Pages: 3,
					Trace: make([]quanyu.TracePoint, 100),
				}, nil
			case 2:
				return &quanyu.BatteryTraceResponse{
					UID:   "UID001",
					Pages: 3,
					Trace: make([]quanyu.TracePoint, 100),
				}, nil
			case 3:
				return &quanyu.BatteryTraceResponse{
					UID:   "UID001",
					Pages: 3,
					Trace: make([]quanyu.TracePoint, 50),
				}, nil
			default:
				t.Fatalf("unexpected page: %d", pageNum)
				return nil, nil
			}
		},
	}

	var storedDocs []storage.BatteryTraceDoc
	store := noopStore()
	store.upsertBatteryTracesFn = func(ctx context.Context, docs []storage.BatteryTraceDoc) (int, error) {
		storedDocs = docs
		return len(docs), nil
	}

	s := newTestSyncer(client, store, &mockDeviceProvider{})
	err := s.syncTraceTask(context.Background(), "UID001")

	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, callPages, "should paginate through all pages")
	assert.Len(t, storedDocs, 250, "should collect all 100+100+50 trace points")
}

func TestSyncTraceTask_SinglePage(t *testing.T) {
	client := &mockQuanyuClient{
		getBatteryTraceFn: func(ctx context.Context, uid, startTime, endTime string, pageNum, pageSize int) (*quanyu.BatteryTraceResponse, error) {
			return &quanyu.BatteryTraceResponse{
				UID:   "UID001",
				Pages: 1,
				Trace: []quanyu.TracePoint{
					{Loc: "loc1", LocTime: "2026-04-13 10:00:00"},
					{Loc: "loc2", LocTime: "2026-04-13 10:01:00"},
				},
			}, nil
		},
	}

	var storedDocs []storage.BatteryTraceDoc
	store := noopStore()
	store.upsertBatteryTracesFn = func(ctx context.Context, docs []storage.BatteryTraceDoc) (int, error) {
		storedDocs = docs
		return len(docs), nil
	}

	s := newTestSyncer(client, store, &mockDeviceProvider{})
	err := s.syncTraceTask(context.Background(), "UID001")

	require.NoError(t, err)
	require.Len(t, storedDocs, 2)
	assert.Equal(t, "loc1", storedDocs[0].Loc)
	assert.Equal(t, "loc2", storedDocs[1].Loc)
	assert.Equal(t, "UID001", storedDocs[0].UID)
}

func TestSyncTraceTask_EmptyResult(t *testing.T) {
	syncTimeUpdated := false
	client := &mockQuanyuClient{
		getBatteryTraceFn: func(ctx context.Context, uid, startTime, endTime string, pageNum, pageSize int) (*quanyu.BatteryTraceResponse, error) {
			return &quanyu.BatteryTraceResponse{UID: "UID001", Pages: 0, Trace: nil}, nil
		},
	}

	store := noopStore()
	store.updateSyncTimeFn = func(ctx context.Context, uid, syncType, syncTime string) error {
		syncTimeUpdated = true
		return nil
	}

	s := newTestSyncer(client, store, &mockDeviceProvider{})
	err := s.syncTraceTask(context.Background(), "UID001")

	require.NoError(t, err)
	assert.True(t, syncTimeUpdated, "sync time should still be updated on empty result")
}

func TestSyncTraceTask_TimeFormatConversion(t *testing.T) {
	var capturedStartTime string

	client := &mockQuanyuClient{
		getBatteryTraceFn: func(ctx context.Context, uid, startTime, endTime string, pageNum, pageSize int) (*quanyu.BatteryTraceResponse, error) {
			if pageNum == 1 {
				capturedStartTime = startTime
			}
			return &quanyu.BatteryTraceResponse{UID: "UID001", Trace: nil}, nil
		},
	}

	store := noopStore()
	store.getSyncStateFn = func(ctx context.Context, uid, syncType string) (*storage.SyncStateDoc, error) {
		return &storage.SyncStateDoc{LastSyncTime: "2026-04-13 10:00:00"}, nil
	}

	s := newTestSyncer(client, store, &mockDeviceProvider{})
	err := s.syncTraceTask(context.Background(), "UID001")

	require.NoError(t, err)
	assert.Equal(t, "20260413100000", capturedStartTime, "time should be converted from datetime to compact format")
}

func TestSyncTraceTask_IncrementalFromState(t *testing.T) {
	var capturedStart, capturedEnd string

	client := &mockQuanyuClient{
		getBatteryTraceFn: func(ctx context.Context, uid, startTime, endTime string, pageNum, pageSize int) (*quanyu.BatteryTraceResponse, error) {
			capturedStart = startTime
			capturedEnd = endTime
			return &quanyu.BatteryTraceResponse{UID: "UID001", Trace: nil}, nil
		},
	}

	store := noopStore()
	store.getSyncStateFn = func(ctx context.Context, uid, syncType string) (*storage.SyncStateDoc, error) {
		return &storage.SyncStateDoc{LastSyncTime: "2026-04-13 10:00:00"}, nil
	}

	s := newTestSyncer(client, store, &mockDeviceProvider{})
	err := s.syncTraceTask(context.Background(), "UID001")

	require.NoError(t, err)
	assert.Equal(t, "20260413100000", capturedStart)
	_, err = time.ParseInLocation("20060102150405", capturedEnd, time.Local)
	assert.NoError(t, err, "endTime should be valid yyyyMMddHHmmss format")
}

func TestSyncTraceTask_APIError(t *testing.T) {
	client := &mockQuanyuClient{
		getBatteryTraceFn: func(ctx context.Context, uid, startTime, endTime string, pageNum, pageSize int) (*quanyu.BatteryTraceResponse, error) {
			return nil, errors.New("network error")
		},
	}

	var errorUpdated bool
	store := noopStore()
	store.updateSyncErrorFn = func(ctx context.Context, uid, syncType, errMsg string) error {
		errorUpdated = true
		assert.Equal(t, "trace", syncType)
		return nil
	}

	s := newTestSyncer(client, store, &mockDeviceProvider{})
	err := s.syncTraceTask(context.Background(), "UID001")

	require.Error(t, err)
	assert.True(t, errorUpdated)
}
