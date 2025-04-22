package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"strings"

	"github.com/dtroode/urlshorter/internal/model"
	"github.com/dtroode/urlshorter/internal/response"
	"github.com/dtroode/urlshorter/internal/storage"
	"github.com/google/uuid"
)

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

func (s *URL) CreateShortURL(ctx context.Context, dto *CreateShortURLDTO) (string, error) {
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

func (s *URL) CreateShortURLBatch(ctx context.Context, dto *CreateShortURLBatchDTO) ([]*response.CreateShortURLBatch, error) {
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

func (s *URL) DeleteURLs(ctx context.Context, dto *DeleteURLsDTO) error {
	doneCh := make(chan struct{})
	// intentionally no defer close doneCh
	// because we need to return nil regardless of deletion result
	// but also want methods to be universal so support for done channel is left for them

	go func() {
		urls, err := s.storage.GetURLs(context.TODO(), dto.ShortKeys)
		if err != nil {
			return
		}

		inputCh := s.urlsGenerator(doneCh, urls)
		userURLsCh := s.filterURLsByUser(doneCh, inputCh, dto.UserID)

		ids := make([]uuid.UUID, 0)
		for url := range userURLsCh {
			ids = append(ids, url.ID)
		}
		err = s.storage.DeleteURLs(context.TODO(), ids)
		if err != nil {
			return
		}
	}()

	return nil
}

func (s *URL) urlsGenerator(doneCh <-chan struct{}, input []*model.URL) <-chan *model.URL {
	inputCh := make(chan *model.URL)

	go func() {
		defer close(inputCh)

		for _, i := range input {
			select {
			case <-doneCh:
				return
			case inputCh <- i:
			}
		}
	}()

	return inputCh
}

func (s *URL) filterURLsByUser(doneCh <-chan struct{}, urlsCh <-chan *model.URL, userID uuid.UUID) <-chan *model.URL {
	userURLs := make(chan *model.URL)

	go func() {
		defer close(userURLs)

		for u := range urlsCh {
			select {
			case <-doneCh:
				return
			default:
				if u.UserID == userID {
					userURLs <- u
				}
			}
		}
	}()

	return userURLs
}
