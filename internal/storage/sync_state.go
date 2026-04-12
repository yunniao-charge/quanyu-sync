package storage

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
)

// SyncStateDoc 同步状态 MongoDB 文档
type SyncStateDoc struct {
	UID            string    `bson:"uid"`
	SyncType       string    `bson:"sync_type"`
	LastSyncTime   string    `bson:"last_sync_time,omitempty"`
	LastSuccessAt  time.Time `bson:"last_success_at,omitempty"`
	SyncCount      int       `bson:"sync_count,omitempty"`
	ErrorCount     int       `bson:"error_count,omitempty"`
	LastError      string    `bson:"last_error,omitempty"`
	LastErrorAt    time.Time `bson:"last_error_at,omitempty"`
}

// GetSyncState 获取指定 uid 和类型的同步状态
func (s *MongoStorage) GetSyncState(ctx context.Context, uid, syncType string) (*SyncStateDoc, error) {
	col := s.Collection("sync_states")
	filter := bson.D{
		{Key: "uid", Value: uid},
		{Key: "sync_type", Value: syncType},
	}

	var doc SyncStateDoc
	err := col.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		return nil, nil // 不存在返回 nil
	}
	return &doc, nil
}

// UpdateSyncTime 更新同步成功时间点
func (s *MongoStorage) UpdateSyncTime(ctx context.Context, uid, syncType, syncTime string) error {
	col := s.Collection("sync_states")
	filter := bson.D{
		{Key: "uid", Value: uid},
		{Key: "sync_type", Value: syncType},
	}
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "last_sync_time", Value: syncTime},
			{Key: "last_success_at", Value: time.Now()},
		}},
		{Key: "$inc", Value: bson.D{
			{Key: "sync_count", Value: 1},
		}},
	}

	opts := options.UpdateOne().SetUpsert(true)
	_, err := col.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		s.logger.Error("更新同步状态失败",
			zap.String("uid", uid),
			zap.String("sync_type", syncType),
			zap.Error(err),
		)
	}
	return err
}

// UpdateSyncError 更新同步错误信息
func (s *MongoStorage) UpdateSyncError(ctx context.Context, uid, syncType, errMsg string) error {
	col := s.Collection("sync_states")
	filter := bson.D{
		{Key: "uid", Value: uid},
		{Key: "sync_type", Value: syncType},
	}
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "last_error", Value: errMsg},
			{Key: "last_error_at", Value: time.Now()},
		}},
		{Key: "$inc", Value: bson.D{
			{Key: "error_count", Value: 1},
			{Key: "sync_count", Value: 1},
		}},
	}

	opts := options.UpdateOne().SetUpsert(true)
	_, err := col.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		s.logger.Error("更新同步错误状态失败",
			zap.String("uid", uid),
			zap.String("sync_type", syncType),
			zap.Error(err),
		)
	}
	return err
}
