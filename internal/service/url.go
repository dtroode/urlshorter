package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"strings"

	"github.com/dtroode/urlshorter/internal/model"
	"github.com/dtroode/urlshorter/internal/request"
	"github.com/dtroode/urlshorter/internal/response"
	"github.com/dtroode/urlshorter/internal/storage"
)

type URLStorage interface {
	GetURL(ctx context.Context, shortKey string) (*model.URL, error)
	SetURL(ctx context.Context, url *model.URL) (*model.URL, error)
	SetURLs(ctx context.Context, urls []*model.URL) (savedURLs []*model.URL, err error)
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

	for range s.shortKeyLength {
		randomIndex := rand.Intn(len(characters))
		randomChar := characters[randomIndex]
		sb.WriteRune(randomChar)
	}

	return sb.String()
}

func (s *URL) GetOriginalURL(ctx context.Context, id string) (string, error) {
	url, err := s.storage.GetURL(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("failed to get original URL: %w", err)
	}

	return url.OriginalURL, nil
}

func (s *URL) CreateShortURL(ctx context.Context, originalURL string) (string, error) {
	shortKey := s.generateString()
	var responseError error

	urlModel := model.NewURL(shortKey, originalURL)
	savedURL, err := s.storage.SetURL(ctx, urlModel)
	if err != nil && !errors.Is(err, storage.ErrConflict) {
		return "", fmt.Errorf("failed to set URL: %w", err)
	}
	shortKey = savedURL.ShortKey

	if errors.Is(err, storage.ErrConflict) {
		responseError = ErrConflict
	}

	shortURL, joinErr := url.JoinPath(s.baseURL, shortKey)
	if joinErr != nil {
		return "", ErrInternal
	}

	return shortURL, responseError
}

func (s *URL) CreateShortURLBatch(ctx context.Context, urls []*request.CreateShortURLBatch) ([]*response.CreateShortURLBatch, error) {
	resp := make([]*response.CreateShortURLBatch, 0)

	urlModels := make([]*model.URL, 0)

	for _, reqURL := range urls {
		shortKey := s.generateString()

		urlModel := model.NewURL(shortKey, reqURL.OriginalURL)
		urlModels = append(urlModels, urlModel)
	}

	savedURLs, err := s.storage.SetURLs(ctx, urlModels)
	if err != nil {
		return nil, fmt.Errorf("failed to set urls: %w", err)
	}

	for _, reqURL := range urls {
		respURL := response.CreateShortURLBatch{
			CorrelationID: reqURL.CorrelationID,
		}

		for _, savedURL := range savedURLs {
			if reqURL.OriginalURL == savedURL.OriginalURL {
				shortURL, err := url.JoinPath(s.baseURL, savedURL.ShortKey)
				if err != nil {
					return nil, ErrInternal
				}

				respURL.ShortURL = shortURL
				resp = append(resp, &respURL)
			}
		}
	}

	return resp, nil
}
