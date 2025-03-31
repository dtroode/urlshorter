package storage

import (
	"context"

	"github.com/dtroode/urlshorter/internal/model"
)

type Storage interface {
	GetURL(ctx context.Context, shortKey string) (*string, error)
	SetURL(ctx context.Context, url *model.URL) error
	Close() error
}
