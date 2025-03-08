package service

import (
	"context"
	"math/rand"
	"net/url"
	"strings"

	internalerror "github.com/dtroode/urlshorter/internal/error"
)

type Storage interface {
	GetURL(ctx context.Context, id string) (*string, error)
	SetURL(ctx context.Context, id, url string) error
}

type URL struct {
	baseURL        string
	shortURLLength int
	storage        Storage
}

func NewURL(
	baseURL string,
	shortURLLength int,
	storage Storage,
) *URL {
	return &URL{
		baseURL:        baseURL,
		shortURLLength: shortURLLength,
		storage:        storage,
	}
}

func (s *URL) generateString() string {
	var characters = []rune("ABCDEF0123456789")
	var sb strings.Builder

	for i := 0; i < s.shortURLLength; i++ {
		randomIndex := rand.Intn(len(characters))
		randomChar := characters[randomIndex]
		sb.WriteRune(randomChar)
	}

	return sb.String()
}

func (s *URL) CreateShortURL(ctx context.Context, originalURL string) (*string, error) {
	shortURL := s.generateString()
	if err := s.storage.SetURL(ctx, shortURL, originalURL); err != nil {
		return nil, err
	}

	fullPath, err := url.JoinPath(s.baseURL, shortURL)
	if err != nil {
		return nil, internalerror.ErrInternal
	}

	return &fullPath, nil
}

func (s *URL) GetOriginalURL(ctx context.Context, id string) (*string, error) {
	url, err := s.storage.GetURL(ctx, id)
	if err != nil {
		return nil, err
	}

	return url, nil
}
