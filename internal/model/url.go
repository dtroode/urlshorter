package model

type URL struct {
	ShortKey    string `json:"short_key"`
	OriginalURL string `json:"original_url"`
}

func NewURL(shortKey, originalURL string) *URL {
	return &URL{
		ShortKey:    shortKey,
		OriginalURL: originalURL,
	}
}
