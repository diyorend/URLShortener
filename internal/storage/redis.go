package storage

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisStorage implements URLStorage using Redis as a cache.
type RedisStorage struct {
	client *redis.Client
}

func NewRedisStorage(ctx context.Context, addr string) (*RedisStorage, error) {
	rdb := redis.NewClient(&redis.Options{Addr: addr})

	for i := 1; i <= 5; i++ {
		if err := rdb.Ping(ctx).Err(); err == nil {
			log.Printf("connected to redis on attempt %d", i)
			return &RedisStorage{client: rdb}, nil
		} else {
			log.Printf("redis attempt %d/5: %v — retrying in 1s", i, err)
			time.Sleep(1 * time.Second)
		}
	}

	return nil, fmt.Errorf("redis: failed to connect")
}

// Save stores a short→long mapping in Redis with a 24-hour TTL.
// The TTL ensures stale cache entries expire naturally.
func (r *RedisStorage) Save(ctx context.Context, short string, long string) error {
	return r.client.Set(ctx, short, long, 24*time.Hour).Err()
}

// Get retrieves from Redis. Returns ErrNotFound on cache miss (redis.Nil).
// This translation is important: CachedStorage checks for ErrNotFound to
// decide whether to fall back to Postgres.
func (r *RedisStorage) Get(ctx context.Context, short string) (string, error) {
	val, err := r.client.Get(ctx, short).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", ErrNotFound // cache miss — not an error
		}
		return "", fmt.Errorf("redis.Get: %w", err)
	}
	return val, nil
}

// IncrementClicks is a no-op in the cache layer.
// Clicks are authoritative in Postgres; Redis only stores URL mappings.
func (r *RedisStorage) IncrementClicks(_ context.Context, _ string) error {
	return nil
}
