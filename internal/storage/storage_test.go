package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	internalerror "github.com/dtroode/urlshorter/internal/error"
)

func TestURL_GetURL(t *testing.T) {
	originalURL := "yandex.ru"

	tests := map[string]struct {
		urlmap        map[string]string
		id            string
		expectedURL   *string
		expectedError error
	}{
		"url not found": {
			urlmap: map[string]string{
				"id1": "yandex.ru",
			},
			id:            "id2",
			expectedError: internalerror.ErrNotFound,
		},
		"url found": {
			urlmap: map[string]string{
				"id1": "yandex.ru",
			},
			id:          "id1",
			expectedURL: &originalURL,
		},
	}

	for tn, tt := range tests {
		t.Run(tn, func(t *testing.T) {
			s := URL{
				urlmap: tt.urlmap,
			}

			resp, err := s.GetURL(context.Background(), tt.id)

			assert.ErrorIs(t, tt.expectedError, err)
			assert.Equal(t, tt.expectedURL, resp)
		})
	}
}

func TestURL_SetURL(t *testing.T) {
	tests := map[string]struct {
		urlmap *map[string]string
		id     string
		url    string
	}{
		"url already exist": {
			urlmap: &map[string]string{
				"id1": "yandex.ru",
			},
			id:  "id2",
			url: "yandex.ru",
		},
		"id already exist": {
			urlmap: &map[string]string{
				"id1": "yandex.ru",
			},
			id:  "id1",
			url: "google.com",
		},
		"new url": {
			urlmap: &map[string]string{},
			id:     "id1",
			url:    "google.com",
		},
	}

	for tn, tt := range tests {
		t.Run(tn, func(t *testing.T) {
			s := URL{
				urlmap: *tt.urlmap,
			}

			err := s.SetURL(context.Background(), tt.id, tt.url)

			require.NoError(t, err)

			assert.Equal(t, tt.url, (*tt.urlmap)[tt.id])
		})
	}
}
