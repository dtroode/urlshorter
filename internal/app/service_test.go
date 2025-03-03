package app

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateShortURL(t *testing.T) {
	tests := map[string]struct {
		originalURL          string
		baseURL              string
		shortURLLength       int
		urlmap               map[string]string
		expectedUrlmapLength int
		expectedLength       int
		expectedError        error
	}{
		// ascii control character used here as base URL
		// this causes url.JoinPath to fail
		"failed to join path": {
			originalURL:    "yandex.ru",
			baseURL:        string(rune(0x7f)),
			shortURLLength: 10,
			urlmap:         make(map[string]string),
			expectedError:  ErrInternal,
		},
		"url already exists": {
			originalURL:    "yandex.ru",
			baseURL:        "http://localhost/",
			shortURLLength: 10,
			urlmap: map[string]string{
				"C69F32242B": "yandex.ru",
			},
			expectedUrlmapLength: 2,
			expectedLength:       27,
		},
		"base url without last slash": {
			originalURL:          "yandex.ru",
			baseURL:              "http://localhost",
			shortURLLength:       10,
			expectedLength:       27,
			urlmap:               make(map[string]string),
			expectedUrlmapLength: 1,
		},
		"base url with last slash": {
			originalURL:          "yandex.ru",
			baseURL:              "http://localhost/",
			shortURLLength:       10,
			expectedLength:       27,
			urlmap:               make(map[string]string),
			expectedUrlmapLength: 1,
		},
	}

	for tn, tt := range tests {
		t.Run(tn, func(t *testing.T) {
			service := Service{
				baseURL:        tt.baseURL,
				shortURLLength: tt.shortURLLength,
				urlmap:         tt.urlmap,
			}

			shortURL, err := service.CreateShortURL(context.Background(), tt.originalURL)

			if tt.expectedError != nil {
				assert.ErrorIs(t, tt.expectedError, err)
			} else {
				require.NoError(t, err)

				assert.Len(t, *shortURL, tt.expectedLength)
				assert.True(t, strings.HasPrefix(*shortURL, tt.baseURL))

				urlParts := strings.Split(*shortURL, "/")
				shortID := urlParts[len(urlParts)-1]
				assert.Equal(t, tt.originalURL, service.urlmap[shortID])
			}

		})
	}
}

func TestGetOriginalURL(t *testing.T) {
	originalURL := "yandex.ru"
	shortKey := "C69F32242B"

	tests := map[string]struct {
		urlmap        map[string]string
		expectedURL   *string
		expectedError error
	}{
		"link not found": {
			urlmap:        make(map[string]string),
			expectedError: ErrNotFound,
		},
		"success": {
			urlmap: map[string]string{
				shortKey: "yandex.ru",
			},
			expectedURL: &originalURL,
		},
	}

	for tn, tt := range tests {
		t.Run(tn, func(t *testing.T) {
			service := Service{
				urlmap:         tt.urlmap,
				shortURLLength: 10,
			}

			originalURL, err := service.GetOriginalURL(context.Background(), shortKey)

			if tt.expectedError != nil {
				assert.ErrorIs(t, tt.expectedError, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedURL, originalURL)
			}
		})
	}
}
