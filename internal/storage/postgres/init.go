package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Initialize(ctx context.Context, db *pgxpool.Pool) error {
	tx, err := db.Begin(ctx)
	defer tx.Rollback(ctx)

	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	if err := createUrlsTable(ctx, tx); err != nil {
		return err
	}

	if err := createUrlsIndexOriginalURL(ctx, tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func createUrlsTable(ctx context.Context, tx pgx.Tx) error {
	sql := `
CREATE TABLE IF NOT EXISTS urls (
id uuid PRIMARY KEY,
short_key varchar(32) NOT NULL UNIQUE,
original_url varchar(256) NOT NULL
);`
	_, err := tx.Exec(ctx, sql)
	if err != nil {
		return fmt.Errorf("failed to create table urls: %w", err)
	}

	return nil
}

func createUrlsIndexOriginalURL(ctx context.Context, tx pgx.Tx) error {
	sql := `CREATE UNIQUE INDEX IF NOT EXISTS urls_original_url_idx ON urls (original_url)`
	_, err := tx.Exec(ctx, sql)
	if err != nil {
		return fmt.Errorf("failed to create index on ulrs (original_url): %w", err)
	}

	return nil
}
