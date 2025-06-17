package dto

import (
	"github.com/google/uuid"

	"github.com/dtroode/urlshorter/internal/request"
)

// CreateShortURL represents a data transfer object for creating a single shortened URL.
// It contains the user ID and original URL needed for URL shortening operations.
type CreateShortURL struct {
	// UserID is the UUID of the user creating the shortened URL.
	UserID uuid.UUID
	// OriginalURL is the full URL to be shortened.
	OriginalURL string
}

// NewCreateShortURL creates a new CreateShortURL DTO instance.
//
// Parameters:
//   - originalURL: The full URL to be shortened
//   - userID: The UUID of the user creating the URL
//
// Returns a pointer to the newly created CreateShortURL instance.
func NewCreateShortURL(originalURL string, userID uuid.UUID) *CreateShortURL {
	return &CreateShortURL{
		UserID:      userID,
		OriginalURL: originalURL,
	}
}

// CreateShortURLBatch represents a data transfer object for batch URL shortening operations.
// It contains a slice of URL requests and the user ID for batch processing.
type CreateShortURLBatch struct {
	// URLs is a slice of URL shortening requests for batch processing.
	URLs []*request.CreateShortURLBatch
	// UserID is the UUID of the user creating the shortened URLs.
	UserID uuid.UUID
}

// NewCreateShortURLBatch creates a new CreateShortURLBatch DTO instance.
//
// Parameters:
//   - urls: A slice of URL shortening requests
//   - userID: The UUID of the user creating the URLs
//
// Returns a pointer to the newly created CreateShortURLBatch instance.
func NewCreateShortURLBatch(urls []*request.CreateShortURLBatch, userID uuid.UUID) *CreateShortURLBatch {
	return &CreateShortURLBatch{
		URLs:   urls,
		UserID: userID,
	}
}

// DeleteURLs represents a data transfer object for URL deletion operations.
// It contains the user ID and a slice of short keys to be deleted.
type DeleteURLs struct {
	// UserID is the UUID of the user deleting the URLs.
	UserID uuid.UUID
	// ShortKeys is a slice of short URL identifiers to be deleted.
	ShortKeys []string
}

// NewDeleteURLs creates a new DeleteURLs DTO instance.
//
// Parameters:
//   - shortKeys: A slice of short URL identifiers to delete
//   - userID: The UUID of the user deleting the URLs
//
// Returns a pointer to the newly created DeleteURLs instance.
func NewDeleteURLs(shortKeys []string, userID uuid.UUID) *DeleteURLs {
	return &DeleteURLs{
		UserID:    userID,
		ShortKeys: shortKeys,
	}
}
