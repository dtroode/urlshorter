package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/dtroode/urlshorter/internal/model"
	"github.com/dtroode/urlshorter/internal/response"
	"github.com/dtroode/urlshorter/internal/service/dto"
	"github.com/dtroode/urlshorter/internal/service/workerpool"
	"github.com/dtroode/urlshorter/internal/storage"
)

const deleteBatchSize = 10

type URLStorage interface {
	GetURL(ctx context.Context, shortKey string) (*model.URL, error)
	GetURLs(ctx context.Context, shortKeys []string) ([]*model.URL, error)
	GetURLsByUserID(ctx context.Context, userID uuid.UUID) ([]*model.URL, error)
	SetURL(ctx context.Context, url *model.URL) (*model.URL, error)
	SetURLs(ctx context.Context, urls []*model.URL) (savedURLs []*model.URL, err error)
	DeleteURLs(ctx context.Context, ids []uuid.UUID) error
}

type URL struct {
	baseURL        string
	shortKeyLength int
	storage        URLStorage
	pool           *workerpool.Pool
}

func NewURL(
	baseURL string,
	shortKeyLength int,
	concurrencyLimit int,
	queueSize int,
	storage URLStorage,
) *URL {
	service := &URL{
		baseURL:        baseURL,
		shortKeyLength: shortKeyLength,
		storage:        storage,
	}

	pool := workerpool.NewPool(concurrencyLimit, queueSize)
	service.pool = pool
	pool.Start()

	return service
}

func (s *URL) generateString() string {
	var characters = []rune("ABCDEF0123456789")
	var sb strings.Builder
	sb.Grow(s.shortKeyLength)

	for range s.shortKeyLength {
		randomIndex := rand.Intn(len(characters))
		randomChar := characters[randomIndex]
		sb.WriteRune(randomChar)
	}

	return sb.String()
}

func (s *URL) GetOriginalURL(ctx context.Context, shortKey string) (string, error) {
	url, err := s.storage.GetURL(ctx, shortKey)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("failed to get original URL: %w", err)
	}

	if url.DeletedAt != nil {
		return "", ErrGone
	}

	return url.OriginalURL, nil
}

func (s *URL) CreateShortURL(ctx context.Context, dto *dto.CreateShortURL) (string, error) {
	shortKey := s.generateString()
	var responseError error

	urlModel := model.NewURL(shortKey, dto.OriginalURL, dto.UserID)
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

func (s *URL) CreateShortURLBatch(ctx context.Context, dto *dto.CreateShortURLBatch) ([]*response.CreateShortURLBatch, error) {
	resp := make([]*response.CreateShortURLBatch, 0)

	urlModels := make([]*model.URL, 0)

	for _, reqURL := range dto.URLs {
		shortKey := s.generateString()

		urlModel := model.NewURL(shortKey, reqURL.OriginalURL, dto.UserID)
		urlModels = append(urlModels, urlModel)
	}

	savedURLs, err := s.storage.SetURLs(ctx, urlModels)
	if err != nil {
		return nil, fmt.Errorf("failed to set urls: %w", err)
	}

	for _, reqURL := range dto.URLs {
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

func (s *URL) GetUserURLs(ctx context.Context, userID uuid.UUID) ([]*response.GetUserURL, error) {
	urls, err := s.storage.GetURLsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get urls: %w", err)
	}

	if len(urls) == 0 {
		return nil, ErrNoContent
	}

	resp := make([]*response.GetUserURL, len(urls))

	for i, u := range urls {
		shortURL, err := url.JoinPath(s.baseURL, u.ShortKey)
		if err != nil {
			return nil, ErrInternal
		}

		respURL := response.GetUserURL{
			ShortURL:    shortURL,
			OriginalURL: u.OriginalURL,
		}
		resp[i] = &respURL
	}

	return resp, nil
}

func (s *URL) DeleteURLs(_ context.Context, data *dto.DeleteURLs) error {
	go func() {
		batches := splitIntoBatches(data.ShortKeys, deleteBatchSize)

		for _, batch := range batches {
			batchDTO := dto.NewDeleteURLs(batch, data.UserID)
			_ = s.pool.Submit(context.Background(), 30*time.Second, s.deleteURLsJob(batchDTO), false)
		}
	}()

	return nil
}

func (s *URL) deleteURLsJob(dto *dto.DeleteURLs) func(context.Context) (any, error) {
	return func(ctx context.Context) (any, error) {
		urls, err := s.storage.GetURLs(ctx, dto.ShortKeys)
		if err != nil {
			return nil, err
		}

		ids := make([]uuid.UUID, 0)
		for _, url := range urls {
			if url.UserID == dto.UserID {
				ids = append(ids, url.ID)
			}
		}

		err = s.storage.DeleteURLs(ctx, ids)
		if err != nil {
			return nil, err
		}

		return nil, nil
	}
}

func splitIntoBatches[T any](items []T, batchSize int) [][]T {
	count := (len(items) + batchSize - 1) / batchSize
	batches := make([][]T, 0, count)

	for batchSize < len(items) {
		batches = append(batches, items[0:batchSize:batchSize])
		items = items[batchSize:]
	}
	if len(items) > 0 {
		batches = append(batches, items)
	}

	return batches
}
