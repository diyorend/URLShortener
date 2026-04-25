package storage

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStorage struct {
	client *redis.Client
}

func NewRedisStorage(ctx context.Context, addr string) (*RedisStorage, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	// Eagerly check connection
	var err error
	for i := 1; i <= 5; i++ {
		if err = rdb.Ping(ctx).Err(); err == nil {
			return &RedisStorage{client: rdb}, nil
		}
		log.Printf("Redis not ready (Attempt %d/5): %v. Retrying...", i, err)
		time.Sleep(1 * time.Second)
	}

	return nil, err
}

func (r *RedisStorage) Save(short string, long string) error {
	// Cache for 24 hours. The Source of Truth is still in Postgres anyway.
	return r.client.Set(context.Background(), short, long, 24*time.Hour).Err()
}

func (r *RedisStorage) Get(short string) (string, error) {
	return r.client.Get(context.Background(), short).Result()
}

func (r *RedisStorage) IncrementClicks(short string) error {
	// We are using Postgres as the source of truth for clicks,
	// so we can leave this empty or use it for Redis-specific metrics later.
	return nil
}
