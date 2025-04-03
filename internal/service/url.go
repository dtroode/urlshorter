package service

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
	"strings"

	internalerror "github.com/dtroode/urlshorter/internal/error"
	"github.com/dtroode/urlshorter/internal/model"
	"github.com/dtroode/urlshorter/internal/request"
	"github.com/dtroode/urlshorter/internal/response"
)

type URLStorage interface {
	GetURL(ctx context.Context, shortKey string) (*model.URL, error)
	SetURL(ctx context.Context, url *model.URL) error
	SetURLs(ctx context.Context, urls []*model.URL) error
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

func (s *URL) GetOriginalURL(ctx context.Context, id string) (*string, error) {
	url, err := s.storage.GetURL(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get original URL: %w", err)
	}

	return &url.OriginalURL, nil
}

func (s *URL) CreateShortURL(ctx context.Context, originalURL string) (*string, error) {
	shortKey := s.generateString()

	urlModel := model.NewURL(shortKey, originalURL)
	if err := s.storage.SetURL(ctx, urlModel); err != nil {
		return nil, fmt.Errorf("failed to set URL: %w", err)
	}

	shortURL, err := url.JoinPath(s.baseURL, shortKey)
	if err != nil {
		return nil, internalerror.ErrInternal
	}

	return &shortURL, nil
}

func (s *URL) CreateShortURLBatch(ctx context.Context, urls []*request.CreateShortURLBatch) ([]*response.CreateShortURLBatch, error) {
	resp := make([]*response.CreateShortURLBatch, 0)

	urlModels := make([]*model.URL, 0)

	for _, reqURL := range urls {
		respURL := response.CreateShortURLBatch{
			CorrelationID: reqURL.CorrelationID,
		}

		shortKey := s.generateString()

		urlModel := model.NewURL(shortKey, reqURL.OriginalURL)
		urlModels = append(urlModels, urlModel)

		shortURL, err := url.JoinPath(s.baseURL, shortKey)
		if err != nil {
			return nil, internalerror.ErrInternal
		}

		respURL.ShortURL = shortURL

		resp = append(resp, &respURL)
	}

	if err := s.storage.SetURLs(ctx, urlModels); err != nil {
		return nil, fmt.Errorf("failed to set urls: %w", err)
	}

	return resp, nil
}
