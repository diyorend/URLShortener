package storage

import (
	"log"
)

type URLStorage interface {
	Save(short string, long string) error
	Get(short string) (string, error)
}

type CachedStorage struct {
	postgres URLStorage
	redis URLStorage
}

func NewCachedStorage(pg URLStorage, rdb URLStorage) *CachedStorage {
	return &CachedStorage{
		postgres: pg,
		redis: rdb,
	}
}

func (c *CachedStorage) Save(short string, long string) error {
	// Save to Postgres first (Source of Truth)
	if err := c.postgres.Save(short, long); err != nil {
		return err
	}
	// Then "warm up" the cache
	return c.redis.Save(short, long)
}

func (c *CachedStorage) Get(short string) (string, error) {
	// 1. Try Redis
	val, err := c.redis.Get(short)
	if err == nil {
		log.Println("Cache Hit: Found in Redis")
		return val, nil
	}

	// 2. Fallback to Postgres
	log.Println("Cache Miss: Checking Postgres")
	val, err = c.postgres.Get(short)
	if err != nil {
		return "", err
	}

	// 3. Re-populate Redis so next time is fast
	_ = c.redis.Save(short, val)
	return val, nil
}


