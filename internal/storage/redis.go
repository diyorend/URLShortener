package storage

import (
	"context"
	"github.com/redis/go-redis/v9"
)

type RedisStorage struct {
	client *redis.Client
}

func NewRedisStorage(addr string) *RedisStorage {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &RedisStorage{client: rdb}
}

// Save implements URLStorage
func (r *RedisStorage) Save(short string, long string) error {
	return r.client.Set(context.Background(), short, long, 0).Err()
}

// Get implements URLStorage
func (r *RedisStorage) Get(short string) (string, error) {
	return r.client.Get(context.Background(), short).Result()
}
