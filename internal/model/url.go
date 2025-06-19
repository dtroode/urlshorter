package model

import (
	"time"

	"github.com/google/uuid"
)

// URL represents a URL entity in the system.
// It contains all the information about a shortened URL including its metadata.
type URL struct {
	// ID is the unique identifier for the URL record.
	// Generated automatically when a new URL is created.
	ID uuid.UUID `json:"id"`

	// ShortKey is the short identifier used in the shortened URL.
	// This is the part that appears after the domain in the shortened URL.
	ShortKey string `json:"short_key"`

	// OriginalURL is the full original URL that was shortened.
	// This is the URL that users will be redirected to.
	OriginalURL string `json:"original_url"`

	// UserID is the identifier of the user who created this URL.
	// Used for ownership and access control.
	UserID uuid.UUID `json:"user_id"`

	// DeletedAt is the timestamp when the URL was marked as deleted.
	// If nil, the URL is active. If not nil, the URL has been soft deleted.
	DeletedAt *time.Time `json:"deleted_at"`
}

// NewURL creates a new URL instance with the provided parameters.
// The ID field is automatically generated using a new UUID.
//
// Parameters:
//   - shortKey: The short identifier for the URL
//   - originalURL: The full original URL to be shortened
//   - userID: The ID of the user creating the URL
//
// Returns a pointer to the newly created URL instance.
func NewURL(shortKey, originalURL string, userID uuid.UUID) *URL {
	return &URL{
		ID:          uuid.New(),
		ShortKey:    shortKey,
		OriginalURL: originalURL,
		UserID:      userID,
	}
}
