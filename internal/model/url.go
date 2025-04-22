package model

import (
	"time"

	"github.com/google/uuid"
)

type URL struct {
	ID          uuid.UUID  `json:"id"`
	ShortKey    string     `json:"short_key"`
	OriginalURL string     `json:"original_url"`
	UserID      uuid.UUID  `json:"user_id"`
	DeletedAt   *time.Time `json:"deleted_at"`
}

func NewURL(shortKey, originalURL string, userID uuid.UUID) *URL {
	return &URL{
		ID:          uuid.New(),
		ShortKey:    shortKey,
		OriginalURL: originalURL,
		UserID:      userID,
	}
}
