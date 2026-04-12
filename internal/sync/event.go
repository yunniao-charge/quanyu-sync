package sync

import (
	"context"
	"time"

	"quanyu-battery-sync/internal/storage"
)

func (s *Syncer) syncEventTask(ctx context.Context, uid string) error {
	state, _ := s.storage.GetSyncState(ctx, uid, "event")

	var startTime string
	if state != nil && state.LastSyncTime != "" {
		startTime = state.LastSyncTime
	} else {
		defaultRange := s.config.Event.DefaultRange
		if defaultRange == 0 {
			defaultRange = 1 * time.Hour
		}
		startTime = time.Now().Add(-defaultRange).Format("2006-01-02 15:04:05")
	}

	endTime := time.Now().Format("2006-01-02 15:04:05")

	var allDocs []storage.BatteryEventDoc
	page := 1
	limit := 100

	for {
		events, err := s.client.GetBatteryEvents(ctx, uid, startTime, endTime, page, limit)
		if err != nil {
			_ = s.storage.UpdateSyncError(ctx, uid, "event", err.Error())
			return err
		}

		for _, item := range events.List {
			allDocs = append(allDocs, storage.BatteryEventDoc{
				UID:      uid,
				Alarm:    item.V2,
				Time:     item.V1,
				V1:       item.V1,
				V2:       item.V2,
				V4:       item.V4,
				V6:       item.V6,
				V7:       item.V7,
				V8:       item.V8,
				V9:       item.V9,
				SyncedAt: time.Now(),
			})
		}

		// 简单分页判断
		totalPages := (events.Total + limit - 1) / limit
		if page >= totalPages || len(events.List) < limit {
			break
		}
		page++
	}

	if len(allDocs) > 0 {
		count, err := s.storage.UpsertBatteryEvents(ctx, allDocs)
		if err != nil {
			_ = s.storage.UpdateSyncError(ctx, uid, "event", err.Error())
			return err
		}
		_ = count
	}

	return s.storage.UpdateSyncTime(ctx, uid, "event", endTime)
}
