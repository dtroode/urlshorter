package response

// CreateShortURL represents a response for a created shortened URL.
// @Description Response structure for a created shortened URL
type CreateShortURL struct {
	// URL is the shortened URL that was created.
	// Contains the full shortened URL including the base URL.
	// @Example "https://shortener.example.com/abc123"
	URL string `json:"result" example:"https://shortener.example.com/abc123"`
}

// CreateShortURLBatch represents a response item for batch URL shortening.
// @Description Response structure for batch URL shortening operations
type CreateShortURLBatch struct {
	// CorrelationID is the unique identifier that was provided in the request.
	// Used to correlate the response with the original request.
	// @Example "req-123"
	CorrelationID string `json:"correlation_id" example:"req-123"`

	// ShortURL is the shortened URL that was created for the original URL.
	// Contains the full shortened URL including the base URL.
	// @Example "https://shortener.example.com/abc123"
	ShortURL string `json:"short_url" example:"https://shortener.example.com/abc123"`
}

// GetUserURL represents a URL entry in the user's URL list.
// @Description Response structure for a user's URL entry
type GetUserURL struct {
	// ShortURL is the shortened URL created by the user.
	// Contains the full shortened URL including the base URL.
	// @Example "https://shortener.example.com/abc123"
	ShortURL string `json:"short_url" example:"https://shortener.example.com/abc123"`

	// OriginalURL is the original URL that was shortened.
	// Contains the full original URL that was provided during creation.
	// @Example "https://example.com/very-long-url-path"
	OriginalURL string `json:"original_url" example:"https://example.com/very-long-url-path"`
}
