package storage

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
)

type PostgresStorage struct {
	conn *pgx.Conn
}

func NewPostgresStorage(ctx context.Context, connStr string) (*PostgresStorage, error) {
	var conn *pgx.Conn
	var err error

	// Retry loop: attempts to connect 5 times with a 2-second delay
	for i := 1; i <= 5; i++ {
		conn, err = pgx.Connect(ctx, connStr)
		if err == nil {
			// Check if the connection is actually alive
			err = conn.Ping(ctx)
			if err == nil {
				log.Printf("Successfully connected to Postgres on attempt %d", i)
				return &PostgresStorage{conn: conn}, nil
			}
		}

		log.Printf("Postgres connection attempt %d failed: %v. Retrying in 2s...", i, err)
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("failed to connect to postgres after 5 attempts: %w", err)
}

// Save implements the URLStorage interface
func (s *PostgresStorage) Save(short string, long string) error {
	_, err := s.conn.Exec(context.Background(),
		"INSERT INTO urls (short_code, long_url) VALUES ($1, $2)", short, long)
	return err
}

func (s *PostgresStorage) Get(short string) (string, error) {
	var long string
	err := s.conn.QueryRow(context.Background(),
		"SELECT long_url FROM urls WHERE short_code = $1", short).Scan(&long)
	return long, err
}

func (s *PostgresStorage) IncrementClicks(short string) error {
	_, err := s.conn.Exec(context.Background(),
		"UPDATE urls SET clicks = clicks + 1 WHERE short_code = $1", short)
	return err
}
