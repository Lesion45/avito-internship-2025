package cache

import (
	"context"
	"time"
)

const Ttl = time.Minute * 30

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}) error
	Del(ctx context.Context, keys ...string) error
	Shutdown() error
}
