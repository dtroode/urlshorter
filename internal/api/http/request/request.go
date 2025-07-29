package request

// CreateShortURL represents a request to create a shortened URL.
// @Description Request structure for creating a shortened URL
type CreateShortURL struct {
	// URL is the original URL to be shortened.
	// Must be a valid HTTP/HTTPS URL.
	// @Example "https://example.com/very-long-url-path"
	URL string `json:"url" example:"https://example.com/very-long-url-path"`
}

// CreateShortURLBatch represents a request item for batch URL shortening.
// @Description Request structure for batch URL shortening operations
type CreateShortURLBatch struct {
	// CorrelationID is a unique identifier for tracking the request in batch operations.
	// Used to correlate the response with the original request.
	// @Example "req-123"
	CorrelationID string `json:"correlation_id" example:"req-123"`

	// OriginalURL is the original URL to be shortened.
	// Must be a valid HTTP/HTTPS URL.
	// @Example "https://example.com/very-long-url-path"
	OriginalURL string `json:"original_url" example:"https://example.com/very-long-url-path"`
}
