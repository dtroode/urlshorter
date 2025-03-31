package storage

import (
	"context"
	"fmt"

	"github.com/dtroode/urlshorter/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	db *pgxpool.Pool
}

func NewDatabase(dsn string) (*Database, error) {
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection pool: %w", err)
	}

	return &Database{
		db: pool,
	}, nil
}
func (s *Database) Close() error {
	s.db.Close()

	return nil
}

func (s *Database) Ping(ctx context.Context) error {
	return s.db.Ping(ctx)
}

func (s *Database) GetURL(ctx context.Context, shortKey string) (*string, error) {
	return nil, nil
}

func (s *Database) SetURL(ctx context.Context, url *model.URL) error {
	return nil
}
