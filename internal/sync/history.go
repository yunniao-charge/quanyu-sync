package sync

import (
	"context"
	"fmt"
	"time"

	"quanyu-battery-sync/internal/storage"
)

func (s *Syncer) syncHistoryDataTask(ctx context.Context, uid string) error {
	state, _ := s.storage.GetSyncState(ctx, uid, "history_data")

	var startTime string
	if state != nil && state.LastSyncTime != "" {
		startTime = state.LastSyncTime
	} else {
		defaultRange := s.config.HistoryData.DefaultRange
		if defaultRange == 0 {
			defaultRange = 1 * time.Hour
		}
		startTime = time.Now().Add(-defaultRange).Format("2006-01-02 15:04:05")
	}

	endTime := time.Now().Format("2006-01-02 15:04:05")

	data, err := s.client.GetBatteryData(ctx, uid, startTime, endTime, 0)
	if err != nil {
		_ = s.storage.UpdateSyncError(ctx, uid, "history_data", err.Error())
		return err
	}

	now := time.Now()
	doc := storage.BatteryHistoryDoc{
		UID:        uid,
		Timestamp:  now,
		TimeScale:  data.TimeScale,
		DeviceBMS1: data.DeviceBMS1,
		DeviceBMS2: data.DeviceBMS2,
		DeviceCT1:  data.DeviceCT1,
		DeviceCT2:  data.DeviceCT2,
		DeviceAV:   data.DeviceAV,
		DeviceCG:   data.DeviceCG,
		DeviceCC:   data.DeviceCC,
		DeviceSA:   data.DeviceSA,
		DeviceREM:  data.DeviceREM,
		VVV:        data.VVV,
		Cap:        data.Cap,
		Rem:        data.Rem,
		SyncedAt:   now,
	}

	count, err := s.storage.UpsertBatteryHistory(ctx, []storage.BatteryHistoryDoc{doc})
	if err != nil {
		_ = s.storage.UpdateSyncError(ctx, uid, "history_data", err.Error())
		return err
	}

	_ = count
	fmt.Printf("history sync: uid=%s, upserted=%d\n", uid, count)

	return s.storage.UpdateSyncTime(ctx, uid, "history_data", endTime)
}
