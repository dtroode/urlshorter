package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/dtroode/urlshorter/database"
	"github.com/dtroode/urlshorter/internal/model"
	"github.com/dtroode/urlshorter/internal/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	db *pgxpool.Pool
}

func NewStorage(dsn string) (*Storage, error) {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection pool: %w", err)
	}

	if err := database.Migrate(ctx, dsn); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return &Storage{
		db: pool,
	}, nil
}

func (s *Storage) Close() error {
	s.db.Close()

	return nil
}

func (s *Storage) Ping(ctx context.Context) error {
	return s.db.Ping(ctx)
}

func (s *Storage) GetURL(ctx context.Context, shortKey string) (*model.URL, error) {
	var url model.URL
	query := `SELECT id, short_key, original_url FROM urls WHERE short_key = $1`
	err := s.db.QueryRow(ctx, query, shortKey).Scan(&url.ID, &url.ShortKey, &url.OriginalURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrNotFound
		}
	}

	return &url, nil
}

func (s *Storage) SetURL(ctx context.Context, url *model.URL) (*model.URL, error) {
	var savedURL model.URL
	query := `
	INSERT INTO urls (id, short_key, original_url) VALUES (@id, @shortKey, @originalURL)
	ON CONFLICT (original_url) DO UPDATE SET short_key = urls.short_key
	RETURNING id, short_key, original_url`
	args := pgx.NamedArgs{
		"id":          url.ID,
		"shortKey":    url.ShortKey,
		"originalURL": url.OriginalURL,
	}
	err := s.db.QueryRow(ctx, query, args).Scan(&savedURL.ID, &savedURL.ShortKey, &savedURL.OriginalURL)
	if err != nil {
		return nil, fmt.Errorf("failed to save url: %w", err)
	}

	if url.ID != savedURL.ID {
		err = storage.ErrConflict
	}

	return &savedURL, err
}

func (s *Storage) SetURLs(ctx context.Context, urls []*model.URL) (savedURLs []*model.URL, err error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
	INSERT INTO urls (id, short_key, original_url) VALUES (@id, @shortKey, @originalURL)
	ON CONFLICT (original_url) DO UPDATE SET short_key = urls.short_key
	RETURNING id, short_key, original_url`

	for _, url := range urls {
		args := pgx.NamedArgs{
			"id":          url.ID,
			"shortKey":    url.ShortKey,
			"originalURL": url.OriginalURL,
		}
		var savedURL model.URL
		err := tx.QueryRow(ctx, query, args).Scan(&savedURL.ID, &savedURL.ShortKey, &savedURL.OriginalURL)
		if err != nil {
			tx.Rollback(ctx)

			return nil, fmt.Errorf("failed to save url: %w", err)
		}

		savedURLs = append(savedURLs, &savedURL)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to commit transcation: %w", err)
	}

	return
}
