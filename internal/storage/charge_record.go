package storage

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// ChargeRecordDoc 充放电记录 MongoDB 文档
type ChargeRecordDoc struct {
	UID          string    `bson:"uid"`
	DeviceID     string    `bson:"device_id,omitempty"`
	ChargeBegin  string    `bson:"charge_begin,omitempty"`
	ChargeEnd    string    `bson:"charge_end,omitempty"`
	BeginSOC     int       `bson:"begin_soc,omitempty"`
	EndSOC       int       `bson:"end_soc,omitempty"`
	ChargeDWh    any       `bson:"charge_dwh,omitempty"`
	ChargeDAh    int       `bson:"charge_dah,omitempty"`
	AccAh        int       `bson:"acc_ah,omitempty"`
	DriveMiles   int       `bson:"drive_miles,omitempty"`
	IDXAuto      int       `bson:"idx_auto"`
	SyncedAt     time.Time `bson:"synced_at"`
}

// UpsertChargeRecords 批量 upsert 充放电记录
func (s *MongoStorage) UpsertChargeRecords(ctx context.Context, docs []ChargeRecordDoc) (int, error) {
	if len(docs) == 0 {
		return 0, nil
	}

	col := s.Collection("charge_records")
	var models []mongo.WriteModel

	for _, doc := range docs {
		filter := bson.D{
			{Key: "uid", Value: doc.UID},
			{Key: "idx_auto", Value: doc.IDXAuto},
		}
		update := bson.D{{Key: "$set", Value: doc}}
		model := mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(true)
		models = append(models, model)
	}

	opts := options.BulkWrite().SetOrdered(false)
	result, err := col.BulkWrite(ctx, models, opts)
	if err != nil {
		return 0, err
	}

	return int(result.UpsertedCount + result.ModifiedCount), nil
}
