package sync

import (
	"context"
	"time"

	"quanyu-battery-sync/internal/storage"
)

func (s *Syncer) syncChargeDataTask(ctx context.Context, uid string) error {
	state, _ := s.storage.GetSyncState(ctx, uid, "charge_data")

	var startTime string
	if state != nil && state.LastSyncTime != "" {
		startTime = state.LastSyncTime
	} else {
		defaultRange := s.config.ChargeData.DefaultRange
		if defaultRange == 0 {
			defaultRange = 2 * time.Hour
		}
		startTime = time.Now().Add(-defaultRange).Format("2006-01-02 15:04:05")
	}

	endTime := time.Now().Format("2006-01-02 15:04:05")

	var allDocs []storage.ChargeRecordDoc
	page := 1
	limit := 100

	for {
		records, err := s.client.GetChargeRecords(ctx, uid, startTime, endTime, page, limit)
		if err != nil {
			_ = s.storage.UpdateSyncError(ctx, uid, "charge_data", err.Error())
			return err
		}

		for _, item := range records.List {
			allDocs = append(allDocs, storage.ChargeRecordDoc{
				UID:         uid,
				DeviceID:    item.DeviceID,
				ChargeBegin: item.ChargeBegin,
				ChargeEnd:   item.ChargeEnd,
				BeginSOC:    item.BeginSOC,
				EndSOC:      item.EndSOC,
				ChargeDWh:   item.ChargeDWh,
				ChargeDAh:   item.ChargeDAh,
				AccAh:       item.AccAh,
				DriveMiles:  item.DriveMiles,
				IDXAuto:     item.IDXAuto,
				SyncedAt:    time.Now(),
			})
		}

		// 检查是否还有下一页
		if page >= records.TotalPage || len(records.List) < limit {
			break
		}
		page++
	}

	if len(allDocs) > 0 {
		count, err := s.storage.UpsertChargeRecords(ctx, allDocs)
		if err != nil {
			_ = s.storage.UpdateSyncError(ctx, uid, "charge_data", err.Error())
			return err
		}
		_ = count
	}

	return s.storage.UpdateSyncTime(ctx, uid, "charge_data", endTime)
}
