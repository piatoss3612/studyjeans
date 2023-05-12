package study

import (
	"context"
	"time"

	"github.com/go-redis/cache/v8"
)

type Cache interface {
	Exists(ctx context.Context, key string) bool
	Get(ctx context.Context, key string, value interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
}

type cacheImpl struct {
	client *cache.Cache
}

func NewCache(client *cache.Cache) Cache {
	return &cacheImpl{client: client}
}

func (c *cacheImpl) Exists(ctx context.Context, key string) bool {
	return c.client.Exists(ctx, key)
}

func (c *cacheImpl) Get(ctx context.Context, key string, value interface{}) error {
	return c.client.Get(ctx, key, value)
}

func (c *cacheImpl) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return c.client.Set(&cache.Item{
		Ctx:   ctx,
		Key:   key,
		Value: value,
		TTL:   ttl,
	})
}
