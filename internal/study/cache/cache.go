package cache

import (
	"context"
	"time"

	"github.com/go-redis/cache/v8"
)

type Cache interface {
	Exists(ctx context.Context, key string) bool
	Get(ctx context.Context, key string, value interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl ...time.Duration) error
}

type redisCache struct {
	client *cache.Cache
}

func NewCache(client *cache.Cache) Cache {
	return &redisCache{client: client}
}

func (c *redisCache) Exists(ctx context.Context, key string) bool {
	return c.client.Exists(ctx, key)
}

func (c *redisCache) Get(ctx context.Context, key string, value interface{}) error {
	return c.client.Get(ctx, key, value)
}

func (c *redisCache) Set(ctx context.Context, key string, value interface{}, ttl ...time.Duration) error {
	item := &cache.Item{
		Ctx:   ctx,
		Key:   key,
		Value: value,
	}

	if len(ttl) > 0 {
		item.TTL = ttl[0]
	}

	return c.client.Set(item)
}
