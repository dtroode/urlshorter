package service

import (
	"github.com/dtroode/urlshorter/internal/request"
	"github.com/google/uuid"
)

type CreateShortURLDTO struct {
	UserID      uuid.UUID
	OriginalURL string
}

func NewCreateShortURLDTO(originalURL string, userID uuid.UUID) *CreateShortURLDTO {
	return &CreateShortURLDTO{
		UserID:      userID,
		OriginalURL: originalURL,
	}
}

type CreateShortURLBatchDTO struct {
	URLs   []*request.CreateShortURLBatch
	UserID uuid.UUID
}

func NewCreateShortURLBatchDTO(urls []*request.CreateShortURLBatch, userID uuid.UUID) *CreateShortURLBatchDTO {
	return &CreateShortURLBatchDTO{
		URLs:   urls,
		UserID: userID,
	}
}

type DeleteURLsDTO struct {
	UserID    uuid.UUID
	ShortKeys []string
}

func NewDeleteURLsDTO(shortKeys []string, userID uuid.UUID) *DeleteURLsDTO {
	return &DeleteURLsDTO{
		UserID:    userID,
		ShortKeys: shortKeys,
	}
}
