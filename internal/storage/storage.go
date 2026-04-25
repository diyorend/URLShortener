package storage

import (
	"context"
	"errors"
	"log"
)

// ErrNotFound is returned when a short code does not exist.
// Callers use errors.Is(err, storage.ErrNotFound) to detect this case.
var ErrNotFound = errors.New("storage: not found")

// URLStorage defines what ANY storage implementation must do.
// Handler and Service depend on this interface — never on a concrete type.
// This is the most important pattern in this codebase.
type URLStorage interface {
	Save(ctx context.Context, short string, long string) error
	Get(ctx context.Context, short string) (string, error)
	IncrementClicks(ctx context.Context, short string) error
}

// CachedStorage is a composite: Redis on top, Postgres underneath.
// It implements URLStorage itself, so the service never knows about the two layers.
type CachedStorage struct {
	postgres URLStorage
	redis    URLStorage
}

func NewCachedStorage(pg URLStorage, rdb URLStorage) *CachedStorage {
	return &CachedStorage{
		postgres: pg,
		redis:    rdb,
	}
}

// Save writes to Postgres first (source of truth), then warms the Redis cache.
func (c *CachedStorage) Save(ctx context.Context, short string, long string) error {
	if err := c.postgres.Save(ctx, short, long); err != nil {
		return err
	}
	// Cache warming: best-effort. If Redis is down, we still succeed.
	if err := c.redis.Save(ctx, short, long); err != nil {
		log.Printf("cache warm failed for %s: %v", short, err)
	}
	return nil
}

// Get tries Redis first (fast path), falls back to Postgres on miss.
func (c *CachedStorage) Get(ctx context.Context, short string) (string, error) {
	val, err := c.redis.Get(ctx, short)
	if err == nil {
		log.Println("cache hit:", short)
		// Don't make the caller wait for the click counter
		go func() {
			if err := c.postgres.IncrementClicks(context.Background(), short); err != nil {
				log.Printf("increment clicks failed for %s: %v", short, err)
			}
		}()
		return val, nil
	}

	log.Println("cache miss:", short)
	val, err = c.postgres.Get(ctx, short)
	if err != nil {
		return "", err // already ErrNotFound if pgx returns pgx.ErrNoRows
	}

	// Re-populate cache; don't block the caller
	go func() {
		if err := c.redis.Save(context.Background(), short, val); err != nil {
			log.Printf("cache re-populate failed for %s: %v", short, err)
		}
		if err := c.postgres.IncrementClicks(context.Background(), short); err != nil {
			log.Printf("increment clicks failed for %s: %v", short, err)
		}
	}()

	return val, nil
}

func (c *CachedStorage) IncrementClicks(ctx context.Context, short string) error {
	return c.postgres.IncrementClicks(ctx, short)
}
