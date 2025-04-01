package model

import (
	"github.com/google/uuid"
)

type URL struct {
	ID          uuid.UUID `json:"id"`
	ShortKey    string    `json:"short_key"`
	OriginalURL string    `json:"original_url"`
}

func NewURL(shortKey, originalURL string) *URL {
	return &URL{
		ID:          uuid.New(),
		ShortKey:    shortKey,
		OriginalURL: originalURL,
	}
}
