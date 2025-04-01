package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Initialize(ctx context.Context, db *pgxpool.Pool) error {
	if err := createUrlsTable(ctx, db); err != nil {
		return fmt.Errorf("failed to initialize tables: %w", err)
	}

	return nil
}

func createUrlsTable(ctx context.Context, db *pgxpool.Pool) error {
	sql := `
CREATE TABLE IF NOT EXISTS urls (
id uuid PRIMARY KEY,
short_key varchar(32) NOT NULL UNIQUE,
original_url varchar(256) NOT NULL
);`
	_, err := db.Exec(ctx, sql)
	if err != nil {
		return fmt.Errorf("failed to create table urls: %w", err)
	}

	return nil
}
