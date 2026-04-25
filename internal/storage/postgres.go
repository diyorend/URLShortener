package storage

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresStorage holds a connection pool (not a single connection).
// A pool handles concurrent requests — a single conn would serialize them.
type PostgresStorage struct {
	pool *pgxpool.Pool
}

func NewPostgresStorage(ctx context.Context, connStr string) (*PostgresStorage, error) {
	var pool *pgxpool.Pool
	var err error

	// Retry loop: Postgres takes a few seconds to start in Docker.
	// Without this, the app crashes on startup before the DB is ready.
	for i := 1; i <= 5; i++ {
		pool, err = pgxpool.New(ctx, connStr)
		if err == nil {
			if err = pool.Ping(ctx); err == nil {
				log.Printf("connected to postgres on attempt %d", i)

				// Configure pool limits — critical under load.
				// Without this, Go opens unlimited connections and Postgres collapses.
				pool.Config().MaxConns = 25
				pool.Config().MinConns = 2

				return &PostgresStorage{pool: pool}, nil
			}
		}
		log.Printf("postgres attempt %d/5 failed: %v — retrying in 2s", i, err)
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("postgres: failed after 5 attempts: %w", err)
}

// Save inserts a new short→long mapping.
func (s *PostgresStorage) Save(ctx context.Context, short string, long string) error {
	_, err := s.pool.Exec(ctx,
		"INSERT INTO urls (short_code, long_url) VALUES ($1, $2)", short, long)
	if err != nil {
		return fmt.Errorf("postgres.Save: %w", err)
	}
	return nil
}

// Get retrieves the long URL for a short code.
// Translates pgx.ErrNoRows → storage.ErrNotFound so callers don't import pgx.
func (s *PostgresStorage) Get(ctx context.Context, short string) (string, error) {
	var long string
	err := s.pool.QueryRow(ctx,
		"SELECT long_url FROM urls WHERE short_code = $1", short).Scan(&long)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFound // convert DB-specific error to our sentinel
		}
		return "", fmt.Errorf("postgres.Get: %w", err)
	}
	return long, nil
}

// IncrementClicks atomically bumps the click counter.
func (s *PostgresStorage) IncrementClicks(ctx context.Context, short string) error {
	_, err := s.pool.Exec(ctx,
		"UPDATE urls SET clicks = clicks + 1 WHERE short_code = $1", short)
	if err != nil {
		return fmt.Errorf("postgres.IncrementClicks: %w", err)
	}
	return nil
}
