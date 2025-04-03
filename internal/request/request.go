package request

type CreateShortURL struct {
	URL string `json:"url"`
}

type CreateShortURLBatch struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}
