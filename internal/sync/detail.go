package sync

import (
	"context"
	"time"

	"quanyu-battery-sync/internal/storage"
)

func (s *Syncer) syncDetailTask(ctx context.Context, uid string) error {
	detail, err := s.client.GetBatteryDetail(ctx, uid)
	if err != nil {
		_ = s.storage.UpdateSyncError(ctx, uid, "detail", err.Error())
		return err
	}

	doc := &storage.BatteryDetailDoc{
		UID:          detail.UID,
		SN:           detail.BTCode,
		SOC:          detail.SOC,
		Voltage:      detail.Voltage,
		OnlineStatus: detail.OnlineStatus,
		DevState:     detail.DevState,
		CellVoltage:  detail.CellVoltage,
		DeviceBMS1:   detail.DeviceBMS1,
		DeviceBMS2:   detail.DeviceBMS2,
		LastPos:      detail.LastPos,
		Remain:       detail.Remain,
		Current:      detail.Current,
		Charge:       detail.Charge,
		Discharge:    detail.Discharge,
		BatTime:      detail.Date,
		TempMOS:      detail.TempMOS,
		TempENV:      detail.TempENV,
		TempC1:       detail.TempC1,
		TempC2:       detail.TempC2,
		Cap:          detail.Cap,
		IMSI:         detail.IMSI,
		IMEI:         detail.IMEI,
		Signal:       detail.Signal,
		DevType:      detail.DevType,
		Loc:          detail.LastPos,
		LocTime:      detail.LocTime,
		N:            detail.N,
		E:            detail.E,
		UpdatedAt:    time.Now(),
		SyncSource:   "pull",
	}

	if err := s.storage.UpsertBatteryDetail(ctx, doc); err != nil {
		_ = s.storage.UpdateSyncError(ctx, uid, "detail", err.Error())
		return err
	}

	return nil
}
