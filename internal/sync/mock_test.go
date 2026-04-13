package sync

import (
	"context"

	"quanyu-battery-sync/internal/quanyu"
	"quanyu-battery-sync/internal/storage"
)

// mockQuanyuClient implements quanyu.QuanyuClient for testing
type mockQuanyuClient struct {
	getBatteryDetailFn  func(ctx context.Context, uid string) (*quanyu.BatteryDetail, error)
	getBatteryDataFn    func(ctx context.Context, uid, startTime, endTime string, last int) (*quanyu.BatteryDataResponse, error)
	getBatteryTraceFn   func(ctx context.Context, uid, startTime, endTime string, pageNum, pageSize int) (*quanyu.BatteryTraceResponse, error)
	getBatteryEventsFn  func(ctx context.Context, uid, startTime, endTime string, page, limit int) (*quanyu.BatteryEventResponse, error)
	getChargeRecordsFn  func(ctx context.Context, uid, beginStart, beginEnd string, page, limit int) (*quanyu.ChargeDataResponse, error)
	subscribeV2Fn       func(ctx context.Context, uid string, list []string, subData []string, notifyURL string) (*quanyu.QuanyuResponse, error)
}

func (m *mockQuanyuClient) GetBatteryDetail(ctx context.Context, uid string) (*quanyu.BatteryDetail, error) {
	return m.getBatteryDetailFn(ctx, uid)
}

func (m *mockQuanyuClient) GetBatteryData(ctx context.Context, uid, startTime, endTime string, last int) (*quanyu.BatteryDataResponse, error) {
	return m.getBatteryDataFn(ctx, uid, startTime, endTime, last)
}

func (m *mockQuanyuClient) GetBatteryTrace(ctx context.Context, uid, startTime, endTime string, pageNum, pageSize int) (*quanyu.BatteryTraceResponse, error) {
	return m.getBatteryTraceFn(ctx, uid, startTime, endTime, pageNum, pageSize)
}

func (m *mockQuanyuClient) GetBatteryEvents(ctx context.Context, uid, startTime, endTime string, page, limit int) (*quanyu.BatteryEventResponse, error) {
	return m.getBatteryEventsFn(ctx, uid, startTime, endTime, page, limit)
}

func (m *mockQuanyuClient) GetChargeRecords(ctx context.Context, uid, beginStart, beginEnd string, page, limit int) (*quanyu.ChargeDataResponse, error) {
	return m.getChargeRecordsFn(ctx, uid, beginStart, beginEnd, page, limit)
}

func (m *mockQuanyuClient) SubscribeV2(ctx context.Context, uid string, list []string, subData []string, notifyURL string) (*quanyu.QuanyuResponse, error) {
	return m.subscribeV2Fn(ctx, uid, list, subData, notifyURL)
}

// mockSyncStorage implements storage.SyncStorage for testing
type mockSyncStorage struct {
	upsertBatteryDetailFn  func(ctx context.Context, doc *storage.BatteryDetailDoc) error
	getBatteryDetailFn     func(ctx context.Context, uid string) (*storage.BatteryDetailDoc, error)
	upsertBatteryHistoryFn func(ctx context.Context, docs []storage.BatteryHistoryDoc) (int, error)
	upsertBatteryTracesFn  func(ctx context.Context, docs []storage.BatteryTraceDoc) (int, error)
	upsertBatteryEventsFn  func(ctx context.Context, docs []storage.BatteryEventDoc) (int, error)
	upsertChargeRecordsFn  func(ctx context.Context, docs []storage.ChargeRecordDoc) (int, error)
	getSyncStateFn         func(ctx context.Context, uid, syncType string) (*storage.SyncStateDoc, error)
	updateSyncTimeFn       func(ctx context.Context, uid, syncType, syncTime string) error
	updateSyncErrorFn      func(ctx context.Context, uid, syncType, errMsg string) error
}

func (m *mockSyncStorage) UpsertBatteryDetail(ctx context.Context, doc *storage.BatteryDetailDoc) error {
	return m.upsertBatteryDetailFn(ctx, doc)
}

func (m *mockSyncStorage) GetBatteryDetail(ctx context.Context, uid string) (*storage.BatteryDetailDoc, error) {
	return m.getBatteryDetailFn(ctx, uid)
}

func (m *mockSyncStorage) UpsertBatteryHistory(ctx context.Context, docs []storage.BatteryHistoryDoc) (int, error) {
	return m.upsertBatteryHistoryFn(ctx, docs)
}

func (m *mockSyncStorage) UpsertBatteryTraces(ctx context.Context, docs []storage.BatteryTraceDoc) (int, error) {
	return m.upsertBatteryTracesFn(ctx, docs)
}

func (m *mockSyncStorage) UpsertBatteryEvents(ctx context.Context, docs []storage.BatteryEventDoc) (int, error) {
	return m.upsertBatteryEventsFn(ctx, docs)
}

func (m *mockSyncStorage) UpsertChargeRecords(ctx context.Context, docs []storage.ChargeRecordDoc) (int, error) {
	return m.upsertChargeRecordsFn(ctx, docs)
}

func (m *mockSyncStorage) GetSyncState(ctx context.Context, uid, syncType string) (*storage.SyncStateDoc, error) {
	return m.getSyncStateFn(ctx, uid, syncType)
}

func (m *mockSyncStorage) UpdateSyncTime(ctx context.Context, uid, syncType, syncTime string) error {
	return m.updateSyncTimeFn(ctx, uid, syncType, syncTime)
}

func (m *mockSyncStorage) UpdateSyncError(ctx context.Context, uid, syncType, errMsg string) error {
	return m.updateSyncErrorFn(ctx, uid, syncType, errMsg)
}

// mockDeviceProvider implements device.DeviceProvider for testing
type mockDeviceProvider struct {
	uids []string
}

func (m *mockDeviceProvider) GetUIDs() []string {
	return m.uids
}

// noopStore returns a mockSyncStorage with all functions set to no-op defaults.
// Tests can override specific fields as needed.
func noopStore() *mockSyncStorage {
	return &mockSyncStorage{
		upsertBatteryDetailFn:  func(ctx context.Context, doc *storage.BatteryDetailDoc) error { return nil },
		getBatteryDetailFn:     func(ctx context.Context, uid string) (*storage.BatteryDetailDoc, error) { return nil, nil },
		upsertBatteryHistoryFn: func(ctx context.Context, docs []storage.BatteryHistoryDoc) (int, error) { return 0, nil },
		upsertBatteryTracesFn:  func(ctx context.Context, docs []storage.BatteryTraceDoc) (int, error) { return 0, nil },
		upsertBatteryEventsFn:  func(ctx context.Context, docs []storage.BatteryEventDoc) (int, error) { return 0, nil },
		upsertChargeRecordsFn:  func(ctx context.Context, docs []storage.ChargeRecordDoc) (int, error) { return 0, nil },
		getSyncStateFn:         func(ctx context.Context, uid, syncType string) (*storage.SyncStateDoc, error) { return nil, nil },
		updateSyncTimeFn:       func(ctx context.Context, uid, syncType, syncTime string) error { return nil },
		updateSyncErrorFn:      func(ctx context.Context, uid, syncType, errMsg string) error { return nil },
	}
}
