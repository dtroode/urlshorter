package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/dtroode/urlshorter/internal/model"
	"github.com/dtroode/urlshorter/internal/storage"
	database "github.com/dtroode/urlshorter/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	db *pgxpool.Pool
}

func NewStorage(dsn string) (*Storage, error) {
	ctx := context.Background()

	// if err := database.Migrate(ctx, dsn); err != nil {
	// 	return nil, fmt.Errorf("failed to initialize database: %w", err)
	// }

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection pool: %w", err)
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
	query := `SELECT id, short_key, original_url FROM urls WHERE short_key = @shortKey`
	args := pgx.NamedArgs{
		"shortKey": shortKey,
	}
	rows, err := s.db.Query(ctx, query, args)
	if err != nil {
		return nil, fmt.Errorf("failed to get rows: %w", err)
	}

	var url model.URL
	rowsProceed := 0
	for rows.Next() {
		if err := rows.Scan(&url.ID, &url.ShortKey, &url.OriginalURL); err != nil {
			return nil, fmt.Errorf("failed to assign database row to model: %w", err)
		}

		rowsProceed++
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if rowsProceed == 0 {
		return nil, storage.ErrNotFound
	}

	return &url, nil
}

func (s *Storage) GetURLByOriginal(ctx context.Context, originalURL string) (*model.URL, error) {
	query := `SELECT id, short_key, original_url FROM urls WHERE original_url = @originalURL LIMIT 1`
	args := pgx.NamedArgs{
		"originalURL": originalURL,
	}
	rows, err := s.db.Query(ctx, query, args)
	if err != nil {
		return nil, fmt.Errorf("failed to get rows: %w", err)
	}

	var url model.URL
	rowsProceed := 0
	for rows.Next() {
		if err := rows.Scan(&url.ID, &url.ShortKey, &url.OriginalURL); err != nil {
			return nil, fmt.Errorf("failed to assign database row to model: %w", err)
		}

		rowsProceed++
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if rowsProceed == 0 {
		return nil, storage.ErrNotFound
	}

	return &url, nil
}

func (s *Storage) SetURL(ctx context.Context, url *model.URL) error {
	query := `
	INSERT INTO urls (id, short_key, original_url) VALUES (@id, @shortKey, @originalURL)
	ON CONFLICT (original_url) DO NOTHING`
	args := pgx.NamedArgs{
		"id":          url.ID,
		"shortKey":    url.ShortKey,
		"originalURL": url.OriginalURL,
	}
	tag, err := s.db.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("failed to save url: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return storage.ErrConflict
	}

	return nil
}

func (s *Storage) SetURLs(ctx context.Context, urls []*model.URL) (savedURLs []*model.URL, err error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
	INSERT INTO urls (id, short_key, original_url) VALUES (@id, @shortKey, @originalURL)
	ON CONFLICT (original_url) DO NOTHING RETURNING id, short_key, original_url`

	for _, url := range urls {
		args := pgx.NamedArgs{
			"id":          url.ID,
			"shortKey":    url.ShortKey,
			"originalURL": url.OriginalURL,
		}
		row := tx.QueryRow(ctx, query, args)

		var savedURL model.URL

		err := row.Scan(&savedURL.ID, &savedURL.ShortKey, &savedURL.OriginalURL)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}
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
