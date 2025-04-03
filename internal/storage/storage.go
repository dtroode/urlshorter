package storage

import (
	"context"

	"github.com/dtroode/urlshorter/internal/model"
)

type Storage interface {
	GetURL(ctx context.Context, shortKey string) (*model.URL, error)
	SetURL(ctx context.Context, url *model.URL) error
	SetURLs(ctx context.Context, urls []*model.URL) error
	Close() error
}
