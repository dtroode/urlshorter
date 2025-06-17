package storage

import (
	"context"

	"github.com/google/uuid"

	"github.com/dtroode/urlshorter/internal/model"
)

// Storage defines the interface for URL storage operations.
type Storage interface {
	Ping(ctx context.Context) error
	GetURL(ctx context.Context, shortKey string) (*model.URL, error)
	GetURLs(ctx context.Context, shortKeys []string) ([]*model.URL, error)
	SetURL(ctx context.Context, url *model.URL) (*model.URL, error)
	SetURLs(ctx context.Context, urls []*model.URL) (savedURLs []*model.URL, err error)
	GetURLsByUserID(ctx context.Context, userID uuid.UUID) ([]*model.URL, error)
	DeleteURLs(ctx context.Context, ids []uuid.UUID) error
	Close() error
}
