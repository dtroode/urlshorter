package service

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	internalerror "github.com/dtroode/urlshorter/internal/error"
	"github.com/dtroode/urlshorter/internal/model"
	"github.com/dtroode/urlshorter/internal/service/mocks"
)

func TestURL_CreateShortURL(t *testing.T) {
	tests := map[string]struct {
		originalURL          string
		baseURL              string
		shortKeyLength       int
		storageError         error
		expectedUrlmapLength int
		expectedLength       int
		expectedError        error
	}{
		"storage error": {
			storageError:  internalerror.ErrInternal,
			expectedError: internalerror.ErrInternal,
		},
		// ascii control character used here as base URL
		// this causes url.JoinPath to fail
		"failed to join path": {
			originalURL:    "yandex.ru",
			baseURL:        string(rune(0x7f)),
			shortKeyLength: 10,
			expectedError:  internalerror.ErrInternal,
		},
		"base url without last slash": {
			originalURL:          "yandex.ru",
			baseURL:              "http://localhost",
			shortKeyLength:       10,
			expectedLength:       27,
			expectedUrlmapLength: 1,
		},
		"base url with last slash": {
			originalURL:          "yandex.ru",
			baseURL:              "http://localhost/",
			shortKeyLength:       10,
			expectedLength:       27,
			expectedUrlmapLength: 1,
		},
	}

	for tn, tt := range tests {
		t.Run(tn, func(t *testing.T) {
			ctx := context.Background()

			ctrl := gomock.NewController(t)

			urlStorage := mocks.NewMockStorage(ctrl)
			urlStorage.EXPECT().SetURL(ctx, gomock.AssignableToTypeOf(&model.URL{})).Return(tt.storageError).Times(1)
			service := URL{
				baseURL:        tt.baseURL,
				shortKeyLength: tt.shortKeyLength,
				storage:        urlStorage,
			}

			shortURL, err := service.CreateShortURL(ctx, tt.originalURL)

			if tt.expectedError != nil {
				assert.ErrorIs(t, tt.expectedError, err)
			} else {
				require.NoError(t, err)

				assert.Len(t, *shortURL, tt.expectedLength)
				assert.True(t, strings.HasPrefix(*shortURL, tt.baseURL))
			}

		})
	}
}

func TestURL_GetOriginalURL(t *testing.T) {
	originalURL := "yandex.ru"
	shortKey := "C69F32242B"

	tests := map[string]struct {
		storageResponse  *string
		storageError     error
		expectedResponse *string
		expectedError    error
	}{
		"storage error": {
			storageError:  internalerror.ErrNotFound,
			expectedError: internalerror.ErrNotFound,
		},
		"success": {
			expectedResponse: &originalURL,
			storageResponse:  &originalURL,
		},
	}

	for tn, tt := range tests {
		t.Run(tn, func(t *testing.T) {
			ctx := context.Background()

			ctrl := gomock.NewController(t)

			urlStorage := mocks.NewMockStorage(ctrl)
			urlStorage.EXPECT().GetURL(ctx, shortKey).Return(tt.storageResponse, tt.storageError).Times(1)
			service := URL{
				shortKeyLength: 10,
				storage:        urlStorage,
			}

			resp, err := service.GetOriginalURL(ctx, shortKey)

			if tt.expectedError != nil {
				assert.ErrorIs(t, tt.expectedError, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResponse, resp)
			}
		})
	}
}
