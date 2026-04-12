package storage

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
)

// BatteryDetailDoc 电池详情 MongoDB 文档
type BatteryDetailDoc struct {
	UID          string      `bson:"uid"`
	SN           string      `bson:"sn,omitempty"`
	SOC          int         `bson:"soc,omitempty"`
	Voltage      string      `bson:"voltage,omitempty"`
	OnlineStatus interface{} `bson:"online_status,omitempty"`
	DevState     int         `bson:"devstate,omitempty"`
	CellVoltage  interface{} `bson:"cell_voltage,omitempty"`
	DeviceBMS1   interface{} `bson:"device_bms1,omitempty"`
	DeviceBMS2   interface{} `bson:"device_bms2,omitempty"`
	LastPos      string      `bson:"last_pos,omitempty"`
	Remain       string      `bson:"remain,omitempty"`
	Current      string      `bson:"current,omitempty"`
	Charge       int         `bson:"charge,omitempty"`
	Discharge    int         `bson:"discharge,omitempty"`
	BatTime      string      `bson:"bat_time,omitempty"`
	TempMOS      int         `bson:"temp_mos,omitempty"`
	TempENV      int         `bson:"temp_env,omitempty"`
	TempC1       int         `bson:"temp_c1,omitempty"`
	TempC2       int         `bson:"temp_c2,omitempty"`
	Cap          string      `bson:"cap,omitempty"`
	IMSI         string      `bson:"imsi,omitempty"`
	IMEI         string      `bson:"imei,omitempty"`
	Signal       interface{} `bson:"signal,omitempty"`
	DevType      int         `bson:"devtype,omitempty"`
	Loc          string      `bson:"loc,omitempty"`
	LocTime      string      `bson:"loc_time,omitempty"`
	N            string      `bson:"n,omitempty"`
	E            string      `bson:"e,omitempty"`
	UpdatedAt    time.Time   `bson:"updated_at"`
	SyncSource   string      `bson:"sync_source"` // pull | callback
}

// UpsertBatteryDetail 更新或插入电池详情（快照）
func (s *MongoStorage) UpsertBatteryDetail(ctx context.Context, doc *BatteryDetailDoc) error {
	col := s.Collection("battery_details")
	filter := bson.D{{Key: "uid", Value: doc.UID}}
	update := bson.D{{Key: "$set", Value: doc}}

	opts := options.UpdateOne().SetUpsert(true)
	_, err := col.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return err
	}
	return nil
}

// UpdateBatteryDetailFromCallback 从回调数据更新电池详情的特定字段
func (s *MongoStorage) UpdateBatteryDetailFromCallback(ctx context.Context, uid string, fields bson.D) error {
	col := s.Collection("battery_details")

	fields = append(fields,
		bson.E{Key: "updated_at", Value: time.Now()},
		bson.E{Key: "sync_source", Value: "callback"},
	)

	filter := bson.D{{Key: "uid", Value: uid}}
	update := bson.D{{Key: "$set", Value: fields}}

	opts := options.UpdateOne().SetUpsert(true)
	_, err := col.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		s.logger.Error("更新电池详情失败",
			zap.String("uid", uid),
			zap.Error(err),
		)
		return err
	}
	return nil
}

// GetBatteryDetail 获取电池详情
func (s *MongoStorage) GetBatteryDetail(ctx context.Context, uid string) (*BatteryDetailDoc, error) {
	col := s.Collection("battery_details")
	filter := bson.D{{Key: "uid", Value: uid}}

	var doc BatteryDetailDoc
	err := col.FindOne(ctx, filter).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &doc, nil
}
