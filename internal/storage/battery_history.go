package storage

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// BatteryHistoryDoc 历史时序数据 MongoDB 文档
type BatteryHistoryDoc struct {
	UID          string      `bson:"uid"`
	Timestamp    time.Time   `bson:"timestamp"`
	TimeScale    interface{} `bson:"time_scale,omitempty"`
	DeviceBMS1   interface{} `bson:"device_bms1,omitempty"`
	DeviceBMS2   interface{} `bson:"device_bms2,omitempty"`
	DeviceCT1    interface{} `bson:"device_ct1,omitempty"`
	DeviceCT2    interface{} `bson:"device_ct2,omitempty"`
	DeviceAV     interface{} `bson:"device_av,omitempty"`
	DeviceCG     interface{} `bson:"device_cg,omitempty"`
	DeviceCC     interface{} `bson:"device_cc,omitempty"`
	DeviceSA     interface{} `bson:"device_sa,omitempty"`
	DeviceREM    interface{} `bson:"device_rem,omitempty"`
	DeviceCOREV  interface{} `bson:"device_core_v,omitempty"`
	VVV          string      `bson:"vvv,omitempty"`
	Cap          string      `bson:"cap,omitempty"`
	Rem          string      `bson:"rem,omitempty"`
	SyncedAt     time.Time   `bson:"synced_at"`
}

// UpsertBatteryHistory 批量 upsert 历史时序数据
func (s *MongoStorage) UpsertBatteryHistory(ctx context.Context, docs []BatteryHistoryDoc) (int, error) {
	if len(docs) == 0 {
		return 0, nil
	}

	col := s.Collection("battery_history")
	var models []mongo.WriteModel

	for _, doc := range docs {
		filter := bson.D{
			{Key: "uid", Value: doc.UID},
			{Key: "timestamp", Value: doc.Timestamp},
		}
		update := bson.D{{Key: "$set", Value: doc}}
		model := mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(true)
		models = append(models, model)
	}

	opts := options.BulkWrite()
	result, err := col.BulkWrite(ctx, models, opts)
	if err != nil {
		return 0, err
	}

	return int(result.UpsertedCount + result.ModifiedCount), nil
}
