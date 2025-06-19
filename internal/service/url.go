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

// URLStorage defines the interface for URL storage operations.
// It provides methods for storing, retrieving, and managing URL entities.
type URLStorage interface {
	// GetURL retrieves a URL by its short key.
	// Returns the URL model or an error if not found.
	GetURL(ctx context.Context, shortKey string) (*model.URL, error)

	// GetURLs retrieves multiple URLs by their short keys.
	// Returns a slice of URL models or an error if retrieval fails.
	GetURLs(ctx context.Context, shortKeys []string) ([]*model.URL, error)

	// GetURLsByUserID retrieves all URLs created by a specific user.
	// Returns a slice of URL models or an error if retrieval fails.
	GetURLsByUserID(ctx context.Context, userID uuid.UUID) ([]*model.URL, error)

	// SetURL stores a single URL in the storage.
	// Returns the saved URL model or an error if storage fails.
	SetURL(ctx context.Context, url *model.URL) (*model.URL, error)

	// SetURLs stores multiple URLs in the storage.
	// Returns a slice of saved URL models or an error if storage fails.
	SetURLs(ctx context.Context, urls []*model.URL) (savedURLs []*model.URL, err error)

	// DeleteURLs marks the specified URLs as deleted.
	// Returns an error if deletion fails.
	DeleteURLs(ctx context.Context, ids []uuid.UUID) error
}

// URL represents the URL shortening service.
// It provides business logic for URL shortening operations including creation,
// retrieval, and management of shortened URLs.
type URL struct {
	// baseURL is the base URL for generating shortened URLs.
	baseURL string
	// shortKeyLength is the length of generated short keys.
	shortKeyLength int
	// storage is the storage interface for URL persistence.
	storage URLStorage
	// pool is the worker pool for background operations.
	pool *workerpool.Pool
}

// NewURL creates a new URL service instance with the provided configuration.
//
// Parameters:
//   - baseURL: The base URL for generating shortened URLs
//   - shortKeyLength: The length of generated short keys
//   - concurrencyLimit: The maximum number of concurrent workers
//   - queueSize: The size of the worker pool queue
//   - storage: The storage implementation for URL persistence
//
// Returns a pointer to the newly created URL service instance.
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

// GetOriginalURL retrieves the original URL associated with a short key.
//
// Parameters:
//   - ctx: The request context
//   - shortKey: The short key to look up
//
// Returns the original URL string or an error if not found or deleted.
// Returns ErrNotFound if the URL doesn't exist.
// Returns ErrGone if the URL has been deleted.
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

// CreateShortURL creates a new shortened URL from the provided DTO.
// Generates a unique short key and stores the URL mapping.
//
// Parameters:
//   - ctx: The request context
//   - dto: The DTO containing the original URL and user ID
//
// Returns the shortened URL string or an error if creation fails.
// Returns ErrConflict if the URL already exists.
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

// CreateShortURLBatch creates multiple shortened URLs in a single operation.
// Processes all URLs in the batch and returns results with correlation IDs.
//
// Parameters:
//   - ctx: The request context
//   - dto: The DTO containing the batch of URLs to shorten
//
// Returns a slice of created URLs with their correlation IDs or an error if creation fails.
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

// GetUserURLs retrieves all URLs created by the specified user.
//
// Parameters:
//   - ctx: The request context
//   - userID: The UUID of the user whose URLs to retrieve
//
// Returns a slice of user URLs or an error if retrieval fails.
// Returns ErrNoContent if the user has no URLs.
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

// DeleteURLs marks the specified URLs as deleted for the given user.
// This operation is performed asynchronously using a worker pool.
//
// Parameters:
//   - ctx: The request context
//   - data: The DTO containing the short keys to delete and user ID
//
// Returns an error if the deletion request fails to be queued.
// The actual deletion is performed asynchronously.
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
