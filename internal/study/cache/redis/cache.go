package redis

import (
	"context"
	"time"

	rcache "github.com/go-redis/cache/v8"
	"github.com/piatoss3612/presentation-helper-bot/internal/study/cache"
)

type redisCache struct {
	client *rcache.Cache
}

func NewCache(client *rcache.Cache) cache.Cache {
	return &redisCache{client: client}
}

func (c *redisCache) Exists(ctx context.Context, key string) bool {
	return c.client.Exists(ctx, key)
}

func (c *redisCache) Get(ctx context.Context, key string, value interface{}) error {
	return c.client.Get(ctx, key, value)
}

func (c *redisCache) Set(ctx context.Context, key string, value interface{}, ttl ...time.Duration) error {
	item := &rcache.Item{
		Ctx:   ctx,
		Key:   key,
		Value: value,
	}

	if len(ttl) > 0 {
		item.TTL = ttl[0]
	}

	return c.client.Set(item)
}
