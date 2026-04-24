package storage

import (
	"context"
	"github.com/jackc/pgx/v5"
)

type PostgresStorage struct {
	conn *pgx.Conn
}

func NewPostgresStorage(ctx context.Context, connStr string) (*PostgresStorage, error) {
	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return nil, err
	}
	return &PostgresStorage{conn: conn}, nil
}

// Save implements the URLStorage interface
func (s *PostgresStorage) Save(short string, long string) error {
	_, err := s.conn.Exec(context.Background(),
	"INSERT INTO urls (short_code, long_url) VALUES ($1, $2)", short, long)
	return err
}

func (s *PostgresStorage) Get (short string) (string, error) {
	var long string
	err := s.conn.QueryRow(context.Background(),
	"SELECT long_url FROM urls WHERE short_code = $1", short).Scan(&long)
	return long, err
}


