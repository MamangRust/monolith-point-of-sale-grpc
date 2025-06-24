package mencache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/MamangRust/monolith-point-of-sale-pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type CacheStore struct {
	ctx    context.Context
	redis  *redis.Client
	logger logger.LoggerInterface
}

func NewCacheStore(ctx context.Context, redis *redis.Client, logger logger.LoggerInterface) *CacheStore {
	return &CacheStore{ctx: ctx, redis: redis, logger: logger}
}

func GetFromCache[T any](store *CacheStore, key string) (*T, bool) {
	cached, err := store.redis.Get(store.ctx, key).Result()
	if err == redis.Nil {
		return nil, false
	}
	if err != nil {
		store.logger.Error("Redis get error", zap.Error(err), zap.String("cacheKey", key))
		return nil, false
	}

	var result T
	if err := json.Unmarshal([]byte(cached), &result); err != nil {
		store.logger.Error("Failed to unmarshal cache", zap.Error(err), zap.String("cacheKey", key))
		return nil, false
	}

	return &result, true
}

func SetToCache[T any](store *CacheStore, key string, data *T, expiration time.Duration) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		store.logger.Error("Failed to marshal cache", zap.Error(err), zap.String("cacheKey", key))
		return
	}

	if err := store.redis.Set(store.ctx, key, jsonData, expiration).Err(); err != nil {
		store.logger.Error("Failed to set cache", zap.Error(err), zap.String("cacheKey", key))
	} else {
		store.logger.Debug("Successfully cached data",
			zap.String("cacheKey", key),
			zap.Duration("expiration", expiration))
	}
}

func DeleteFromCache(store *CacheStore, key string) {
	if err := store.redis.Del(store.ctx, key).Err(); err != nil {
		store.logger.Error("Failed to delete cache", zap.Error(err), zap.String("cacheKey", key))
	}
}
