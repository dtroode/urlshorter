package app

import (
	"context"
	"math/rand"
	"net/url"
	"strings"
)

type Service struct {
	baseURL        string
	shortURLLength int
	urlmap         map[string]string
}

func NewService(baseURL string, shortURLLength int) *Service {
	return &Service{
		baseURL:        baseURL,
		shortURLLength: shortURLLength,
		urlmap:         make(map[string]string),
	}
}

func (s *Service) generateString() string {
	var characters = []rune("ABCDEF0123456789")
	var sb strings.Builder

	for i := 0; i < s.shortURLLength; i++ {
		randomIndex := rand.Intn(len(characters))
		randomChar := characters[randomIndex]
		sb.WriteRune(randomChar)
	}

	return sb.String()
}

func (s *Service) CreateShortURL(ctx context.Context, originalURL string) (*string, error) {
	shortURL := s.generateString()
	s.urlmap[shortURL] = originalURL

	fullPath, err := url.JoinPath(s.baseURL, shortURL)
	if err != nil {
		return nil, ErrInternal
	}

	return &fullPath, nil
}

func (s *Service) GetOriginalURL(ctx context.Context, shortUrl string) (*string, error) {
	val, ok := s.urlmap[shortUrl]

	if !ok {
		return nil, ErrNotFound
	}

	return &val, nil
}
