package storage

import (
	"context"

	"github.com/dtroode/urlshorter/internal/model"
)

type Storage interface {
	Ping(ctx context.Context) error
	GetURL(ctx context.Context, shortKey string) (*model.URL, error)
	SetURL(ctx context.Context, url *model.URL) (*model.URL, error)
	SetURLs(ctx context.Context, urls []*model.URL) (savedURLs []*model.URL, err error)
	Close() error
}
