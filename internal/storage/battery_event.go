package storage

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// BatteryEventDoc 事件数据 MongoDB 文档
type BatteryEventDoc struct {
	UID      string    `bson:"uid"`
	Alarm    string    `bson:"alarm,omitempty"`
	Type     int       `bson:"type,omitempty"`
	Time     string    `bson:"time"`
	V1       string    `bson:"v1,omitempty"`
	V2       string    `bson:"v2,omitempty"`
	V4       string    `bson:"v4,omitempty"`
	V6       any       `bson:"v6,omitempty"`
	V7       any       `bson:"v7,omitempty"`
	V8       string    `bson:"v8,omitempty"`
	V9       string    `bson:"v9,omitempty"`
	SyncedAt time.Time `bson:"synced_at"`
}

// UpsertBatteryEvents 批量 upsert 事件数据
func (s *MongoStorage) UpsertBatteryEvents(ctx context.Context, docs []BatteryEventDoc) (int, error) {
	if len(docs) == 0 {
		return 0, nil
	}

	col := s.Collection("battery_events")
	var models []mongo.WriteModel

	for _, doc := range docs {
		filter := bson.D{
			{Key: "uid", Value: doc.UID},
			{Key: "alarm", Value: doc.Alarm},
			{Key: "time", Value: doc.Time},
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
