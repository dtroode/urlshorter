package storage

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	internalerror "github.com/dtroode/urlshorter/internal/error"
)

func TestInMemory_GetURL(t *testing.T) {
	originalURL := "yandex.ru"

	tests := map[string]struct {
		urlmap        URLMap
		id            string
		expectedURL   *string
		expectedError error
	}{
		"url not found": {
			urlmap: URLMap{
				"id1": &URLData{
					ShortURL:    "id1",
					OriginalURL: "yandex.ru",
				},
			},
			id:            "id2",
			expectedError: internalerror.ErrNotFound,
		},
		"url found": {
			urlmap: URLMap{
				"id1": &URLData{
					ShortURL:    "id1",
					OriginalURL: "yandex.ru",
				},
			},
			id:          "id1",
			expectedURL: &originalURL,
		},
	}

	for tn, tt := range tests {
		t.Run(tn, func(t *testing.T) {
			s := InMemory{
				urlmap: tt.urlmap,
			}

			resp, err := s.GetURL(context.Background(), tt.id)

			assert.ErrorIs(t, tt.expectedError, err)
			assert.Equal(t, tt.expectedURL, resp)
		})
	}
}

type dummyFile struct {
	*bufio.Writer
}

func (f *dummyFile) Close() error {
	return nil
}

func TestURL_SetURL(t *testing.T) {
	buf := bytes.NewBuffer(nil)

	tests := map[string]struct {
		urlmap URLMap
		id     string
		url    string
	}{
		"url already exist": {
			urlmap: URLMap{
				"id1": &URLData{
					ShortURL:    "id1",
					OriginalURL: "yandex.ru",
				},
			},
			id:  "id2",
			url: "yandex.ru",
		},
		"id already exist": {
			urlmap: URLMap{
				"id1": &URLData{
					ShortURL:    "id1",
					OriginalURL: "yandex.ru",
				},
			},
			id:  "id1",
			url: "google.com",
		},
		"new url": {
			urlmap: URLMap{},
			id:     "id1",
			url:    "google.com",
		},
	}

	for tn, tt := range tests {
		t.Run(tn, func(t *testing.T) {
			s := InMemory{
				urlmap:  tt.urlmap,
				file:    &dummyFile{Writer: bufio.NewWriter(buf)},
				encoder: json.NewEncoder(buf),
			}

			expectedData := URLData{
				ShortURL:    tt.id,
				OriginalURL: tt.url,
			}

			err := s.SetURL(context.Background(), tt.id, tt.url)
			require.NoError(t, err)

			assert.Equal(t, &expectedData, (tt.urlmap)[tt.id])

			line, err := buf.ReadBytes('\n')
			require.NoError(t, err)

			writtenData := URLData{}
			err = json.Unmarshal(line, &writtenData)
			require.NoError(t, err)

			assert.Equal(t, expectedData, writtenData)
		})
	}
}
