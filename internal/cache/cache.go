package cache

import (
	"context"
	"time"
)

type Cache interface {
	Exists(ctx context.Context, key string) bool
	Get(ctx context.Context, key string, value interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl ...time.Duration) error
}
