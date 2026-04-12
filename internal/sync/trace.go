package sync

import (
	"context"
	"time"

	"quanyu-battery-sync/internal/storage"
)

func (s *Syncer) syncTraceTask(ctx context.Context, uid string) error {
	state, _ := s.storage.GetSyncState(ctx, uid, "trace")

	var startTime string
	if state != nil && state.LastSyncTime != "" {
		startTime = state.LastSyncTime
	} else {
		defaultRange := s.config.Trace.DefaultRange
		if defaultRange == 0 {
			defaultRange = 1 * time.Hour
		}
		startTime = time.Now().Add(-defaultRange).Format("20060102150405")
	}

	// trace API 的时间格式是 yyyyMMddHHmmss，但 state 存的是 yyyy-MM-dd HH:mm:ss
	// 需要转换
	traceStart := startTime
	if len(startTime) == 19 {
		// "2006-01-02 15:04:05" -> "20060102150405"
		t, err := time.Parse("2006-01-02 15:04:05", startTime)
		if err == nil {
			traceStart = t.Format("20060102150405")
		}
	}

	traceEnd := time.Now().Format("20060102150405")

	var allDocs []storage.BatteryTraceDoc
	page := 1
	pageSize := 100

	for {
		trace, err := s.client.GetBatteryTrace(ctx, uid, traceStart, traceEnd, page, pageSize)
		if err != nil {
			_ = s.storage.UpdateSyncError(ctx, uid, "trace", err.Error())
			return err
		}

		for _, point := range trace.Trace {
			allDocs = append(allDocs, storage.BatteryTraceDoc{
				UID:      uid,
				Loc:      point.Loc,
				LocTime:  point.LocTime,
				SyncedAt: time.Now(),
			})
		}

		// 检查是否还有下一页
		if page >= trace.Pages || len(trace.Trace) < pageSize {
			break
		}
		page++
	}

	if len(allDocs) > 0 {
		count, err := s.storage.UpsertBatteryTraces(ctx, allDocs)
		if err != nil {
			_ = s.storage.UpdateSyncError(ctx, uid, "trace", err.Error())
			return err
		}
		_ = count
	}

	endTime := time.Now().Format("2006-01-02 15:04:05")
	return s.storage.UpdateSyncTime(ctx, uid, "trace", endTime)
}
