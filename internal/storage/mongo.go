package storage

import (
	"context"
	"fmt"
	"time"

	"quanyu-battery-sync/internal/config"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
)

// MongoStorage MongoDB 存储管理器
type MongoStorage struct {
	client   *mongo.Client
	db       *mongo.Database
	logger   *zap.Logger
	dbName   string
}

// NewMongoStorage 创建 MongoDB 存储实例
func NewMongoStorage(cfg config.MongoDBConfig, logger *zap.Logger) (*MongoStorage, error) {
	opts := options.Client().ApplyURI(cfg.URI)
	if cfg.Username != "" && cfg.Password != "" {
		opts.SetAuth(options.Credential{
			Username: cfg.Username,
			Password: cfg.Password,
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, fmt.Errorf("连接 MongoDB 失败: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("Ping MongoDB 失败: %w", err)
	}

	db := client.Database(cfg.Database)

	s := &MongoStorage{
		client: client,
		db:     db,
		logger: logger,
		dbName: cfg.Database,
	}

	return s, nil
}

// EnsureIndexes 确保所有集合的索引存在
func (s *MongoStorage) EnsureIndexes(ctx context.Context) error {
	indexes := []struct {
		collection string
		models     []mongo.IndexModel
	}{
		{
			collection: "battery_details",
			models: []mongo.IndexModel{
				{Keys: bson.D{{Key: "uid", Value: 1}}, Options: options.Index().SetUnique(true)},
				{Keys: bson.D{{Key: "updated_at", Value: -1}}},
			},
		},
		{
			collection: "battery_history",
			models: []mongo.IndexModel{
				{Keys: bson.D{{Key: "uid", Value: 1}, {Key: "timestamp", Value: 1}}, Options: options.Index().SetUnique(true)},
				{Keys: bson.D{{Key: "uid", Value: 1}, {Key: "synced_at", Value: -1}}},
			},
		},
		{
			collection: "battery_traces",
			models: []mongo.IndexModel{
				{Keys: bson.D{{Key: "uid", Value: 1}, {Key: "loc_time", Value: 1}}, Options: options.Index().SetUnique(true)},
				{Keys: bson.D{{Key: "uid", Value: 1}, {Key: "synced_at", Value: -1}}},
			},
		},
		{
			collection: "battery_events",
			models: []mongo.IndexModel{
				{Keys: bson.D{{Key: "uid", Value: 1}, {Key: "alarm", Value: 1}, {Key: "time", Value: 1}}, Options: options.Index().SetUnique(true)},
				{Keys: bson.D{{Key: "uid", Value: 1}, {Key: "synced_at", Value: -1}}},
			},
		},
		{
			collection: "charge_records",
			models: []mongo.IndexModel{
				{Keys: bson.D{{Key: "uid", Value: 1}, {Key: "idx_auto", Value: 1}}, Options: options.Index().SetUnique(true)},
				{Keys: bson.D{{Key: "uid", Value: 1}, {Key: "charge_begin", Value: -1}}},
			},
		},
		{
			collection: "callback_alarms",
			models: []mongo.IndexModel{
				{Keys: bson.D{{Key: "uid", Value: 1}, {Key: "alarm", Value: 1}, {Key: "time", Value: 1}}, Options: options.Index().SetUnique(true)},
				{Keys: bson.D{{Key: "received_at", Value: -1}}},
			},
		},
		{
			collection: "callback_online",
			models: []mongo.IndexModel{
				{Keys: bson.D{{Key: "uid", Value: 1}, {Key: "time", Value: 1}}, Options: options.Index().SetUnique(true)},
				{Keys: bson.D{{Key: "received_at", Value: -1}}},
			},
		},
		{
			collection: "sync_states",
			models: []mongo.IndexModel{
				{Keys: bson.D{{Key: "uid", Value: 1}, {Key: "sync_type", Value: 1}}, Options: options.Index().SetUnique(true)},
				{Keys: bson.D{{Key: "sync_type", Value: 1}, {Key: "last_sync_time", Value: 1}}},
			},
		},
	}

	for _, idx := range indexes {
		col := s.db.Collection(idx.collection)
		_, err := col.Indexes().CreateMany(ctx, idx.models)
		if err != nil {
			s.logger.Error("创建索引失败",
				zap.String("collection", idx.collection),
				zap.Error(err),
			)
			return fmt.Errorf("创建集合 %s 索引失败: %w", idx.collection, err)
		}
		s.logger.Info("索引创建完成", zap.String("collection", idx.collection))
	}

	return nil
}

// Collection 获取指定集合
func (s *MongoStorage) Collection(name string) *mongo.Collection {
	return s.db.Collection(name)
}

// Close 关闭连接
func (s *MongoStorage) Close(ctx context.Context) error {
	return s.client.Disconnect(ctx)
}
