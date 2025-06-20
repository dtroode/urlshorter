package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dtroode/urlshorter/database"
	"github.com/dtroode/urlshorter/internal/model"
	"github.com/dtroode/urlshorter/internal/storage"
)

// Storage represents PostgreSQL storage implementation.
type Storage struct {
	db *pgxpool.Pool
}

// NewStorage creates new PostgreSQL storage instance.
func NewStorage(dsn string) (*Storage, error) {
	ctx := context.Background()

	conf, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse postgres dsn: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, conf)
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

// Close closes the storage and database connection pool.
func (s *Storage) Close() error {
	s.db.Close()

	return nil
}

// Ping checks if the database is available.
func (s *Storage) Ping(ctx context.Context) error {
	return s.db.Ping(ctx)
}

// GetURL retrieves a URL by its short key.
func (s *Storage) GetURL(ctx context.Context, shortKey string) (*model.URL, error) {
	var url model.URL
	query := `SELECT id, short_key, original_url, user_id, deleted_at FROM urls WHERE short_key = $1`
	err := s.db.QueryRow(ctx, query, shortKey).Scan(&url.ID, &url.ShortKey, &url.OriginalURL, &url.UserID, &url.DeletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get url: %w", err)
	}

	return &url, nil
}

// GetURLs retrieves multiple URLs by their short keys.
func (s *Storage) GetURLs(ctx context.Context, shortKeys []string) ([]*model.URL, error) {
	query := `SELECT id, short_key, original_url, user_id, deleted_at FROM urls WHERE short_key = ANY ($1)`

	keys := &pgtype.TextArray{}
	keys.Set(shortKeys)
	rows, err := s.db.Query(ctx, query, keys)
	if err != nil {
		return nil, fmt.Errorf("failed to query rows: %w", err)
	}
	defer rows.Close()

	urls := make([]*model.URL, 0)

	for rows.Next() {
		var url model.URL
		err := rows.Scan(&url.ID, &url.ShortKey, &url.OriginalURL, &url.UserID, &url.DeletedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		urls = append(urls, &url)
	}

	return urls, nil
}

// GetURLsByUserID retrieves all URLs created by a specific user.
func (s *Storage) GetURLsByUserID(ctx context.Context, userID uuid.UUID) ([]*model.URL, error) {
	query := `SELECT id, short_key, original_url, user_id, deleted_at FROM urls WHERE user_id = $1 AND deleted_at IS NULL`
	rows, err := s.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query rows: %w", err)
	}
	defer rows.Close()

	urls := make([]*model.URL, 0)

	for rows.Next() {
		var url model.URL
		err := rows.Scan(&url.ID, &url.ShortKey, &url.OriginalURL, &url.UserID, &url.DeletedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		urls = append(urls, &url)
	}

	return urls, nil
}

// SetURL stores a single URL in the storage.
func (s *Storage) SetURL(ctx context.Context, url *model.URL) (*model.URL, error) {
	var savedURL model.URL
	query := `
	INSERT INTO urls (id, short_key, original_url, user_id) VALUES (@id, @shortKey, @originalURL, @userID)
	ON CONFLICT (original_url) DO UPDATE SET short_key = urls.short_key
	RETURNING id, short_key, original_url, user_id`
	args := pgx.NamedArgs{
		"id":          url.ID,
		"shortKey":    url.ShortKey,
		"originalURL": url.OriginalURL,
		"userID":      url.UserID,
	}
	err := s.db.QueryRow(ctx, query, args).Scan(&savedURL.ID, &savedURL.ShortKey, &savedURL.OriginalURL, &savedURL.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to save url: %w", err)
	}

	if url.ID != savedURL.ID {
		err = storage.ErrConflict
	}

	return &savedURL, err
}

// SetURLs stores multiple URLs in the storage.
func (s *Storage) SetURLs(ctx context.Context, urls []*model.URL) (savedURLs []*model.URL, err error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
	INSERT INTO urls (id, short_key, original_url, user_id) VALUES (@id, @shortKey, @originalURL, @userID)
	ON CONFLICT (original_url) DO UPDATE SET short_key = urls.short_key
	RETURNING id, short_key, original_url, user_id`

	for _, url := range urls {
		args := pgx.NamedArgs{
			"id":          url.ID,
			"shortKey":    url.ShortKey,
			"originalURL": url.OriginalURL,
			"userID":      url.UserID,
		}
		var savedURL model.URL
		err := tx.QueryRow(ctx, query, args).Scan(&savedURL.ID, &savedURL.ShortKey, &savedURL.OriginalURL, &savedURL.UserID)
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

// DeleteURLs marks the specified URLs as deleted.
func (s *Storage) DeleteURLs(ctx context.Context, ids []uuid.UUID) error {
	query := `UPDATE urls SET deleted_at = now() WHERE id = ANY($1)`
	_, err := s.db.Exec(ctx, query, ids)
	if err != nil {
		return fmt.Errorf("failed to exec query: %w", err)
	}

	return nil
}
