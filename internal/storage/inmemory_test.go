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
	"github.com/dtroode/urlshorter/internal/model"
)

func TestInMemory_GetURL(t *testing.T) {
	originalURL := "yandex.ru"

	tests := map[string]struct {
		urlmap        URLMap
		id            string
		expectedURL   string
		expectedError error
	}{
		"url not found": {
			urlmap: URLMap{
				"id1": &model.URL{
					ShortKey:    "id1",
					OriginalURL: "yandex.ru",
				},
			},
			id:            "id2",
			expectedError: internalerror.ErrNotFound,
		},
		"url found": {
			urlmap: URLMap{
				"id1": &model.URL{
					ShortKey:    "id1",
					OriginalURL: "yandex.ru",
				},
			},
			id:          "id1",
			expectedURL: originalURL,
		},
	}

	for tn, tt := range tests {
		t.Run(tn, func(t *testing.T) {
			s := InMemory{
				urlmap: tt.urlmap,
			}

			resp, err := s.GetURL(context.Background(), tt.id)

			assert.ErrorIs(t, tt.expectedError, err)
			if tt.expectedError == nil {
				assert.Equal(t, tt.expectedURL, resp.OriginalURL)
			}
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
		url    *model.URL
	}{
		"url already exist": {
			urlmap: URLMap{
				"id1": &model.URL{
					ShortKey:    "abcd1",
					OriginalURL: "yandex.ru",
				},
			},
			url: &model.URL{
				ShortKey:    "abcd2",
				OriginalURL: "yandex.ru",
			},
		},
		"id already exist": {
			urlmap: URLMap{
				"id1": &model.URL{
					ShortKey:    "abcd1",
					OriginalURL: "yandex.ru",
				},
			},
			url: &model.URL{
				ShortKey:    "abcd2",
				OriginalURL: "yandex.ru",
			},
		},
		"new url": {
			urlmap: URLMap{},
			url: &model.URL{
				ShortKey:    "abcd2",
				OriginalURL: "yandex.ru",
			},
		},
	}

	for tn, tt := range tests {
		t.Run(tn, func(t *testing.T) {
			s := InMemory{
				urlmap:  tt.urlmap,
				file:    &dummyFile{Writer: bufio.NewWriter(buf)},
				encoder: json.NewEncoder(buf),
			}

			err := s.SetURL(context.Background(), tt.url)
			require.NoError(t, err)

			assert.Equal(t, tt.url, (tt.urlmap)[tt.url.ShortKey])

			line, err := buf.ReadBytes('\n')
			require.NoError(t, err)

			writtenData := &model.URL{}
			err = json.Unmarshal(line, writtenData)
			require.NoError(t, err)

			assert.Equal(t, tt.url, writtenData)
		})
	}
}
