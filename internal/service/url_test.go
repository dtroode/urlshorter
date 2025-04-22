package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/dtroode/urlshorter/internal/model"
	"github.com/dtroode/urlshorter/internal/request"
	"github.com/dtroode/urlshorter/internal/response"
	"github.com/dtroode/urlshorter/internal/service/mocks"
	"github.com/dtroode/urlshorter/internal/storage"
)

func TestURL_GetOriginalURL(t *testing.T) {
	originalURL := "yandex.ru"
	shortKey := "C69F32242B"
	deletedAt := time.Now()

	tests := map[string]struct {
		shortKey         string
		storageResponse  *model.URL
		storageError     error
		expectedResponse string
		expectedError    error
	}{
		"storage error": {
			shortKey:      shortKey,
			storageError:  errors.New("storage error"),
			expectedError: fmt.Errorf("failed to get original URL: %w", errors.New("storage error")),
		},
		"does not exist": {
			shortKey:      "does not exist",
			storageError:  storage.ErrNotFound,
			expectedError: ErrNotFound,
		},
		"deleted": {
			shortKey: shortKey,
			storageResponse: &model.URL{
				ID:          uuid.New(),
				ShortKey:    shortKey,
				OriginalURL: originalURL,
				DeletedAt:   &deletedAt,
			},
			expectedError: ErrGone,
		},
		"success": {
			shortKey: shortKey,
			storageResponse: &model.URL{
				ID:          uuid.New(),
				ShortKey:    shortKey,
				OriginalURL: originalURL,
				DeletedAt:   nil,
			},
			expectedResponse: originalURL,
		},
	}

	for tn, tt := range tests {
		t.Run(tn, func(t *testing.T) {
			ctx := context.Background()

			urlStorage := mocks.NewURLStorage(t)
			urlStorage.On("GetURL", ctx, shortKey).Once().Return(tt.storageResponse, tt.storageError)
			service := URL{
				shortKeyLength: 10,
				storage:        urlStorage,
			}

			resp, err := service.GetOriginalURL(ctx, shortKey)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResponse, resp)
			}
		})
	}
}

func TestURL_CreateShortURL(t *testing.T) {
	userID := uuid.New()

	tests := map[string]struct {
		originalURL          string
		baseURL              string
		shortKeyLength       int
		setURLResponse       *model.URL
		setURLError          error
		expectedUrlmapLength int
		expectedLength       int
		expectedError        error
	}{
		"storage error": {
			originalURL:   "yandex.ru",
			setURLError:   errors.New("storage error"),
			expectedError: fmt.Errorf("failed to set URL: %w", errors.New("storage error")),
		},
		// ascii control character used here as base URL
		// this causes url.JoinPath to fail
		"failed to join path": {
			originalURL: "yandex.ru",
			baseURL:     string(rune(0x7f)),
			setURLResponse: &model.URL{
				OriginalURL: "yandex.ru",
				ShortKey:    "ABCDE",
			},
			shortKeyLength: 10,
			expectedError:  ErrInternal,
		},
		"url already exists": {
			originalURL: "yandex.ru",
			baseURL:     "http://localhost",
			setURLResponse: &model.URL{
				OriginalURL: "yandex.ru",
				ShortKey:    "ABCDE",
			},
			setURLError:          storage.ErrConflict,
			shortKeyLength:       5,
			expectedLength:       22,
			expectedUrlmapLength: 1,
			expectedError:        storage.ErrConflict,
		},
		"base url without last slash": {
			originalURL: "yandex.ru",
			baseURL:     "http://localhost",
			setURLResponse: &model.URL{
				OriginalURL: "yandex.ru",
				ShortKey:    "ABCDE",
			},
			shortKeyLength:       10,
			expectedLength:       22,
			expectedUrlmapLength: 1,
		},
		"base url with last slash": {
			originalURL: "yandex.ru",
			baseURL:     "http://localhost/",
			setURLResponse: &model.URL{
				OriginalURL: "yandex.ru",
				ShortKey:    "ABCDE",
			},
			shortKeyLength:       10,
			expectedLength:       22,
			expectedUrlmapLength: 1,
		},
	}

	for tn, tt := range tests {
		t.Run(tn, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			urlStorage := mocks.NewURLStorage(t)
			urlStorage.On("SetURL", mock.Anything, mock.Anything).Once().Return(tt.setURLResponse, tt.setURLError)
			service := URL{
				baseURL:        tt.baseURL,
				shortKeyLength: tt.shortKeyLength,
				storage:        urlStorage,
			}

			dto := NewCreateShortURLDTO(tt.originalURL, userID)
			shortURL, err := service.CreateShortURL(ctx, dto)
			assert.Equal(t, tt.expectedError, err)

			if tt.expectedError == nil {
				assert.Equal(t, tt.expectedError, err)

				urlStorage.AssertCalled(t, "SetURL", ctx,
					mock.MatchedBy(func(url *model.URL) bool {
						return url.OriginalURL == tt.originalURL && url.UserID == userID
					}))

				assert.Len(t, shortURL, tt.expectedLength)
				assert.True(t, strings.HasPrefix(shortURL, tt.baseURL))
			}

		})
	}
}

func TestURL_CreateShortURLBatch(t *testing.T) {
	userID := uuid.New()

	tests := map[string]struct {
		originalURLs          []*request.CreateShortURLBatch
		baseURL               string
		shortKeyLength        int
		savedURLs             []*model.URL
		setURLsError          error
		existingURL           *model.URL
		getURLByOriginalError error
		expectedURLs          []*response.CreateShortURLBatch
		expectedLength        int
		expectedError         error
	}{
		"save urls error": {
			originalURLs: []*request.CreateShortURLBatch{
				{
					CorrelationID: "1",
					OriginalURL:   "yandex.ru",
				},
				{
					CorrelationID: "2",
					OriginalURL:   "google.com",
				},
			},
			baseURL:       "http://localhost/",
			setURLsError:  errors.New("storage error"),
			expectedError: fmt.Errorf("failed to set urls: %w", errors.New("storage error")),
		},
		"save urls success, already exists": {
			originalURLs: []*request.CreateShortURLBatch{
				{
					CorrelationID: "1",
					OriginalURL:   "yandex.ru",
				},
			},
			baseURL: "http://localhost/",
			savedURLs: []*model.URL{
				{
					OriginalURL: "yandex.ru",
					ShortKey:    "ABOBA",
				},
			},
			existingURL: &model.URL{
				OriginalURL: "yandex.ru",
				ShortKey:    "ABOBA",
			},
			shortKeyLength: 5,
			expectedLength: 22,
			expectedURLs: []*response.CreateShortURLBatch{
				{
					CorrelationID: "1",
					ShortURL:      "http://localhost/ABOBA",
				},
			},
		},
		"save urls success, all new": {
			originalURLs: []*request.CreateShortURLBatch{
				{
					CorrelationID: "1",
					OriginalURL:   "yandex.ru",
				},
			},
			baseURL: "http://localhost/",
			savedURLs: []*model.URL{
				{
					OriginalURL: "yandex.ru",
					ShortKey:    "ABCDE",
				},
			},
			shortKeyLength: 5,
			expectedLength: 22,
			expectedURLs: []*response.CreateShortURLBatch{
				{
					CorrelationID: "1",
					ShortURL:      "http://localhost/ABCDE",
				},
			},
		},
		// ascii control character used here as base URL
		// this causes url.JoinPath to fail
		"failed to join path": {
			originalURLs: []*request.CreateShortURLBatch{
				{
					CorrelationID: "1",
					OriginalURL:   "yandex.ru",
				},
				{
					CorrelationID: "2",
					OriginalURL:   "google.com",
				},
			},
			baseURL: string(rune(0x7f)),
			savedURLs: []*model.URL{
				{
					OriginalURL: "yandex.ru",
					ShortKey:    "ABCDE",
				},
			},
			shortKeyLength: 5,
			expectedError:  ErrInternal,
		},

		"base url without last slash": {
			originalURLs: []*request.CreateShortURLBatch{
				{
					CorrelationID: "1",
					OriginalURL:   "yandex.ru",
				},
			},
			baseURL: "http://localhost",
			savedURLs: []*model.URL{
				{
					OriginalURL: "yandex.ru",
					ShortKey:    "ABCDE",
				},
			},
			shortKeyLength: 10,
			expectedLength: 27,
			expectedURLs: []*response.CreateShortURLBatch{
				{
					CorrelationID: "1",
					ShortURL:      "http://localhost/ABCDE",
				},
			},
		},
	}

	for tn, tt := range tests {
		t.Run(tn, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			urlStorage := mocks.NewURLStorage(t)
			urlStorage.On("SetURLs", mock.Anything, mock.Anything).Maybe().Return(tt.savedURLs, tt.setURLsError)
			urlStorage.On("GetURLByOriginal", mock.Anything, mock.Anything).Maybe().Return(tt.existingURL, tt.getURLByOriginalError)
			service := URL{
				baseURL:        tt.baseURL,
				shortKeyLength: tt.shortKeyLength,
				storage:        urlStorage,
			}

			dto := NewCreateShortURLBatchDTO(tt.originalURLs, userID)
			shortURLs, err := service.CreateShortURLBatch(ctx, dto)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
			} else {
				require.NoError(t, err)

				require.Len(t, shortURLs, len(tt.originalURLs))

				require.Equal(t, tt.expectedURLs, shortURLs)

				shortKeys := make([]string, len(shortURLs))
				for i, respURL := range shortURLs {
					urlParts := strings.Split(respURL.ShortURL, "/")
					shortKey := urlParts[len(urlParts)-1]
					shortKeys[i] = shortKey
				}

				urlStorage.AssertCalled(t, "SetURLs", ctx,
					mock.MatchedBy(func(urls []*model.URL) bool {
						for i, u := range urls {
							return u.OriginalURL == tt.originalURLs[i].OriginalURL && u.UserID == userID
						}
						return true
					}))
			}
		})
	}
}

func TestURL_GetUserURLs(t *testing.T) {
	userID := uuid.New()

	tests := map[string]struct {
		baseURL          string
		storageResponse  []*model.URL
		storageError     error
		expectedResponse []*response.GetUserURL
		expectedError    error
	}{
		"storage error": {
			baseURL:       "http://localhost",
			storageError:  errors.New("storage error"),
			expectedError: fmt.Errorf("failed to get urls: %w", errors.New("storage error")),
		},
		"no urls": {
			baseURL:         "http://localhost",
			storageResponse: make([]*model.URL, 0),
			expectedError:   ErrNoContent,
		},
		// ascii control character used here as base URL
		// this causes url.JoinPath to fail
		"failed to join path": {
			baseURL: string(rune(0x7f)),
			storageResponse: []*model.URL{
				{
					ID:          uuid.New(),
					OriginalURL: "http://yandex.ru",
					ShortKey:    "ABCDE",
					UserID:      userID,
				},
				{
					ID:          uuid.New(),
					OriginalURL: "http://google.com",
					ShortKey:    "ABOBA",
					UserID:      userID,
				},
			},
			expectedError: ErrInternal,
		},
		"success base without last slash": {
			baseURL: "http://localhost",
			storageResponse: []*model.URL{
				{
					ID:          uuid.New(),
					OriginalURL: "http://yandex.ru",
					ShortKey:    "ABCDE",
					UserID:      userID,
				},
				{
					ID:          uuid.New(),
					OriginalURL: "http://google.com",
					ShortKey:    "ABOBA",
					UserID:      userID,
				},
			},
			expectedResponse: []*response.GetUserURL{
				{
					ShortURL:    "http://localhost/ABCDE",
					OriginalURL: "http://yandex.ru",
				},
				{
					ShortURL:    "http://localhost/ABOBA",
					OriginalURL: "http://google.com",
				},
			},
		},
		"success base with last slash": {
			baseURL: "http://localhost/",
			storageResponse: []*model.URL{
				{
					ID:          uuid.New(),
					OriginalURL: "http://yandex.ru",
					ShortKey:    "ABCDE",
					UserID:      userID,
				},
				{
					ID:          uuid.New(),
					OriginalURL: "http://google.com",
					ShortKey:    "ABOBA",
					UserID:      userID,
				},
			},
			expectedResponse: []*response.GetUserURL{
				{
					ShortURL:    "http://localhost/ABCDE",
					OriginalURL: "http://yandex.ru",
				},
				{
					ShortURL:    "http://localhost/ABOBA",
					OriginalURL: "http://google.com",
				},
			},
		},
	}

	for tn, tt := range tests {
		t.Run(tn, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			urlStorage := mocks.NewURLStorage(t)
			urlStorage.On("GetURLsByUserID", ctx, userID).Once().
				Return(tt.storageResponse, tt.storageError)

			service := URL{
				baseURL:        tt.baseURL,
				shortKeyLength: 5,
				storage:        urlStorage,
			}

			urls, err := service.GetUserURLs(ctx, userID)

			assert.Equal(t, tt.expectedResponse, urls)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}

func TestURL_DeleteURLs(t *testing.T) {
	userID := uuid.New()
	shortKeys := []string{"ggl", "ydx"}
	urls := []*model.URL{
		{
			ID:          uuid.New(),
			ShortKey:    "ggl",
			OriginalURL: "https://google.com",
			UserID:      userID,
		},
		{
			ID:          uuid.New(),
			ShortKey:    "ydx",
			OriginalURL: "https://ya.ru",
			UserID:      uuid.New(),
		},
	}
	dto := NewDeleteURLsDTO(shortKeys, userID)

	t.Run("get urls error", func(t *testing.T) {
		t.Parallel()

		urlStorage := mocks.NewURLStorage(t)

		wg := sync.WaitGroup{}
		wg.Add(1)

		urlStorage.On("GetURLs", context.TODO(), shortKeys).Once().
			Run(func(args mock.Arguments) {
				defer wg.Done()
			}).
			Return(nil, errors.New("service error"))

		service := NewURL("base", 3, urlStorage)
		service.DeleteURLs(context.Background(), dto)

		wg.Wait()
		urlStorage.AssertNotCalled(t, "DeleteURLs")
		urlStorage.AssertExpectations(t)
	})

	t.Run("get urls success", func(t *testing.T) {
		t.Parallel()

		wg := sync.WaitGroup{}
		wg.Add(2)

		urlStorage := mocks.NewURLStorage(t)
		urlStorage.On("GetURLs", context.TODO(), shortKeys).Once().
			Run(func(args mock.Arguments) {
				defer wg.Done()
			}).
			Return(urls, nil)
		urlStorage.On("DeleteURLs", context.TODO(), []uuid.UUID{urls[0].ID}).Once().
			Run(func(args mock.Arguments) {
				defer wg.Done()
			}).
			Return(nil)

		service := NewURL("base", 3, urlStorage)
		service.DeleteURLs(context.Background(), dto)

		wg.Wait()
		urlStorage.AssertExpectations(t)
	})
}
