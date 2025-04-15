package storage

import (
	"context"

	"github.com/dtroode/urlshorter/internal/model"
	"github.com/google/uuid"
)

type Storage interface {
	Ping(ctx context.Context) error
	GetURL(ctx context.Context, shortKey string) (*model.URL, error)
	SetURL(ctx context.Context, url *model.URL) (*model.URL, error)
	SetURLs(ctx context.Context, urls []*model.URL) (savedURLs []*model.URL, err error)
	GetURLByUserID(ctx context.Context, userID uuid.UUID) ([]*model.URL, error)
	Close() error
}
