package redis

import (
	"context"
	"encoding/json"

	"avito-internship/internal/cache"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RedisCache struct {
	Client *redis.Client
	Log    *zap.Logger
}

func NewRedisCache(url string, log *zap.Logger) *RedisCache {
	opts, err := redis.ParseURL(url)
	if err != nil {
		panic(err)
	}

	client := redis.NewClient(opts)

	return &RedisCache{
		Client: client,
		Log:    log,
	}
}

func (s *RedisCache) Get(ctx context.Context, key string) (string, error) {
	return s.Client.Get(ctx, key).Result()
}

func (s *RedisCache) Set(ctx context.Context, key string, value interface{}) error {
	valueBytes, err := json.Marshal(value)
	if err != nil {
		s.Log.Error("Failed to marshal value for caching",
			zap.String("key", key),
			zap.Error(err),
		)
		return err
	}

	return s.Client.Set(ctx, key, valueBytes, cache.Ttl).Err()
}

func (s *RedisCache) Del(ctx context.Context, keys ...string) error {
	return s.Client.Del(ctx, keys...).Err()
}

func (s *RedisCache) Shutdown() error {
	if s.Client != nil {
		err := s.Client.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
