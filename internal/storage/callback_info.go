package storage

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// CallbackInfoDoc info 回调推送数据 MongoDB 文档
type CallbackInfoDoc struct {
	UID         string    `bson:"uid"`
	DevType     int       `bson:"devtype,omitempty"`
	SN          string    `bson:"sn,omitempty"`
	Loc         string    `bson:"loc,omitempty"`
	Remain      int       `bson:"remain,omitempty"`
	Online      int       `bson:"online,omitempty"`
	Voltage     int       `bson:"voltage,omitempty"`
	Charge      int       `bson:"charge,omitempty"`
	Discharge   int       `bson:"discharge,omitempty"`
	BatTime     string    `bson:"bat_time,omitempty"`
	ReceivedAt  time.Time `bson:"received_at"`
	AppID       string    `bson:"appid,omitempty"`
}

// UpsertCallbackInfo 更新或插入 info 回调数据（快照）
func (s *MongoStorage) UpsertCallbackInfo(ctx context.Context, doc *CallbackInfoDoc) error {
	col := s.Collection("callback_info")
	filter := bson.D{{Key: "uid", Value: doc.UID}}
	update := bson.D{{Key: "$set", Value: doc}}

	opts := options.UpdateOne().SetUpsert(true)
	_, err := col.UpdateOne(ctx, filter, update, opts)
	return err
}
