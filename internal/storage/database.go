package storage

import (
	"context"
	"fmt"

	"github.com/dtroode/urlshorter/internal/model"
	"github.com/dtroode/urlshorter/internal/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	db *pgxpool.Pool
}

func NewDatabase(dsn string) (*Database, error) {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection pool: %w", err)
	}

	if err := postgres.Initialize(ctx, pool); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
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

func (s *Database) GetURL(ctx context.Context, shortKey string) (*model.URL, error) {
	query := `SELECT id, short_key, original_url FROM urls WHERE short_key = @shortKey`
	args := pgx.NamedArgs{
		"shortKey": shortKey,
	}
	row := s.db.QueryRow(ctx, query, args)

	var url model.URL
	if err := row.Scan(&url.ID, &url.ShortKey, &url.OriginalURL); err != nil {
		return nil, fmt.Errorf("failed to assign database row to model: %w", err)
	}

	return &url, nil
}

func (s *Database) SetURL(ctx context.Context, url *model.URL) error {
	query := `INSERT INTO urls (id, short_key, original_url) VALUES (@id, @shortKey, @originalURL)`
	args := pgx.NamedArgs{
		"id":          url.ID,
		"shortKey":    url.ShortKey,
		"originalURL": url.OriginalURL,
	}
	_, err := s.db.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("failed to save url: %w", err)
	}

	return nil
}

func (s *Database) SetURLs(ctx context.Context, urls []*model.URL) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to bgin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `INSERT INTO urls (id, short_key, original_url) VALUES (@id, @shortKey, @originalURL)`

	for _, url := range urls {
		args := pgx.NamedArgs{
			"id":          url.ID,
			"shortKey":    url.ShortKey,
			"originalURL": url.OriginalURL,
		}
		// pgx automatically prepares statement by default
		_, err := tx.Exec(ctx, query, args)
		if err != nil {
			tx.Rollback(ctx)

			return fmt.Errorf("failed to save url: %w", err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transcation: %w", err)
	}

	return nil
}
