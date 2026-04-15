package storage

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// CallbackAlarmDoc alarm 回调推送数据 MongoDB 文档
type CallbackAlarmDoc struct {
	UID        string    `bson:"uid"`
	Alarm      string    `bson:"alarm"`
	AlarmData  string    `bson:"alarm_data"`
	Time       string    `bson:"time"`
	ReceivedAt time.Time `bson:"received_at"`
	AppID      string    `bson:"appid,omitempty"`
}

// InsertCallbackAlarm 插入 alarm 回调数据
func (s *MongoStorage) InsertCallbackAlarm(ctx context.Context, doc *CallbackAlarmDoc) error {
	col := s.Collection("callback_alarms")
	// 使用 upsert 去重
	filter := bson.D{
		{Key: "uid", Value: doc.UID},
		{Key: "alarm", Value: doc.Alarm},
		{Key: "time", Value: doc.Time},
	}
	update := bson.D{{Key: "$set", Value: doc}}
	opts := options.UpdateOne().SetUpsert(true)
	_, err := col.UpdateOne(ctx, filter, update, opts)
	return err
}
