package model

import (
	"time"

	"github.com/google/uuid"
)

// URL represents a URL entity in the system.
// It contains all the information about a shortened URL including its metadata.
type URL struct {
	// DeletedAt is the timestamp when the URL was marked as deleted.
	// If nil, the URL is active. If not nil, the URL has been soft deleted.
	DeletedAt *time.Time
	// ShortKey is the short identifier used in the shortened URL.
	// This is the part that appears after the domain in the shortened URL.
	ShortKey string
	// OriginalURL is the full original URL that was shortened.
	// This is the URL that users will be redirected to.
	OriginalURL string
	// ID is the unique identifier for the URL record.
	// Generated automatically when a new URL is created.
	ID uuid.UUID
	// UserID is the identifier of the user who created this URL.
	// Used for ownership and access control.
	UserID uuid.UUID
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

// CreateShortURLParams represents the parameters for creating a short URL.
type CreateShortURLParams struct {
	// OriginalURL is the full URL to be shortened.
	OriginalURL string
	// UserID is the UUID of the user creating the shortened URL.
	UserID uuid.UUID
}

// GetUserURL represents a URL returned to a user.
type GetUserURL struct {
	// ShortURL is the shortened URL that was created by the user.
	ShortURL string
	// OriginalURL is the original URL that was shortened.
	OriginalURL string
}

// URLWithCorrelation represents a URL with a correlation ID.
type URLWithCorrelation struct {
	// CorrelationID is a unique identifier for tracking the request in batch operations.
	CorrelationID string
	// OriginalURL is the original URL to be shortened.
	OriginalURL string
}

// CreateShortURLBatchParams represents the parameters for creating a batch of short URLs.
type CreateShortURLBatchParams struct {
	// URLs is a slice of URL shortening requests for batch processing.
	URLs []URLWithCorrelation
	// UserID is the UUID of the user creating the shortened URLs.
	UserID uuid.UUID
}

// CreateShortURLBatch represents a single URL in a batch creation response.
type CreateShortURLBatch struct {
	CorrelationID string
	ShortURL      string
}

// DeleteURLsParams represents the parameters for deleting a batch of short URLs.
type DeleteURLsParams struct {
	// ShortKeys is a slice of short URL identifiers to be deleted.
	ShortKeys []string
	// UserID is the UUID of the user deleting the URLs.
	UserID uuid.UUID
}

// GetStats represents the statistics of the service.
type GetStats struct {
	// URLs is the total number of shortened URLs in the service.
	URLs int64
	// Users is the total number of unique users in the service.
	Users int64
}
