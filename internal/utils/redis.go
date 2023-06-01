package utils

import (
	"context"
	"time"

	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
)

func ConnectRedisCache(ctx context.Context, addr string, ttl time.Duration) (*cache.Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return cache.New(&cache.Options{
		Redis:      client,
		LocalCache: cache.NewTinyLFU(1000, ttl),
	}), nil
}
