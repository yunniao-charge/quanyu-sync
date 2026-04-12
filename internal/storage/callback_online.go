package storage

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// CallbackOnlineDoc online 回调推送数据 MongoDB 文档
type CallbackOnlineDoc struct {
	UID        string    `bson:"uid"`
	Online     int       `bson:"online,omitempty"`
	Time       string    `bson:"time"`
	ReceivedAt time.Time `bson:"received_at"`
	AppID      string    `bson:"appid,omitempty"`
}

// InsertCallbackOnline 插入 online 回调数据
func (s *MongoStorage) InsertCallbackOnline(ctx context.Context, doc *CallbackOnlineDoc) error {
	col := s.Collection("callback_online")
	filter := bson.D{
		{Key: "uid", Value: doc.UID},
		{Key: "time", Value: doc.Time},
	}
	update := bson.D{{Key: "$set", Value: doc}}
	opts := options.UpdateOne().SetUpsert(true)
	_, err := col.UpdateOne(ctx, filter, update, opts)
	return err
}
