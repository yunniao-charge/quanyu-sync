package storage

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// BatteryTraceDoc 轨迹数据 MongoDB 文档
type BatteryTraceDoc struct {
	UID      string    `bson:"uid"`
	Loc      string    `bson:"loc,omitempty"`
	LocTime  string    `bson:"loc_time"`
	SyncedAt time.Time `bson:"synced_at"`
}

// UpsertBatteryTraces 批量 upsert 轨迹数据
func (s *MongoStorage) UpsertBatteryTraces(ctx context.Context, docs []BatteryTraceDoc) (int, error) {
	if len(docs) == 0 {
		return 0, nil
	}

	col := s.Collection("battery_traces")
	var models []mongo.WriteModel

	for _, doc := range docs {
		filter := bson.D{
			{Key: "uid", Value: doc.UID},
			{Key: "loc_time", Value: doc.LocTime},
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
