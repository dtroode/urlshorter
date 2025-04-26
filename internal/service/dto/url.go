package dto

import (
	"github.com/dtroode/urlshorter/internal/request"
	"github.com/google/uuid"
)

type CreateShortURL struct {
	UserID      uuid.UUID
	OriginalURL string
}

func NewCreateShortURL(originalURL string, userID uuid.UUID) *CreateShortURL {
	return &CreateShortURL{
		UserID:      userID,
		OriginalURL: originalURL,
	}
}

type CreateShortURLBatch struct {
	URLs   []*request.CreateShortURLBatch
	UserID uuid.UUID
}

func NewCreateShortURLBatch(urls []*request.CreateShortURLBatch, userID uuid.UUID) *CreateShortURLBatch {
	return &CreateShortURLBatch{
		URLs:   urls,
		UserID: userID,
	}
}

type DeleteURLs struct {
	UserID    uuid.UUID
	ShortKeys []string
}

func NewDeleteURLs(shortKeys []string, userID uuid.UUID) *DeleteURLs {
	return &DeleteURLs{
		UserID:    userID,
		ShortKeys: shortKeys,
	}
}
