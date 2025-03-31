package service

import (
	"context"
	"math/rand"
	"net/url"
	"strings"

	internalerror "github.com/dtroode/urlshorter/internal/error"
	"github.com/dtroode/urlshorter/internal/model"
)

type URLStorage interface {
	GetURL(ctx context.Context, shortKey string) (*string, error)
	SetURL(ctx context.Context, url *model.URL) error
}

type URL struct {
	baseURL        string
	shortKeyLength int
	storage        URLStorage
}

func NewURL(
	baseURL string,
	shortKeyLength int,
	storage URLStorage,
) *URL {
	return &URL{
		baseURL:        baseURL,
		shortKeyLength: shortKeyLength,
		storage:        storage,
	}
}

func (s *URL) generateString() string {
	var characters = []rune("ABCDEF0123456789")
	var sb strings.Builder

	for i := 0; i < s.shortKeyLength; i++ {
		randomIndex := rand.Intn(len(characters))
		randomChar := characters[randomIndex]
		sb.WriteRune(randomChar)
	}

	return sb.String()
}

func (s *URL) CreateShortURL(ctx context.Context, originalURL string) (*string, error) {
	shortKey := s.generateString()

	urlModel := model.NewURL(shortKey, originalURL)
	if err := s.storage.SetURL(ctx, urlModel); err != nil {
		return nil, err
	}

	shortURL, err := url.JoinPath(s.baseURL, shortKey)
	if err != nil {
		return nil, internalerror.ErrInternal
	}

	return &shortURL, nil
}

func (s *URL) GetOriginalURL(ctx context.Context, id string) (*string, error) {
	url, err := s.storage.GetURL(ctx, id)
	if err != nil {
		return nil, err
	}

	return url, nil
}
