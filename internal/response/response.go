package response

type CreateShortURL struct {
	URL string `json:"result"`
}

type CreateShortURLBatch struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
