package inmemory

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dtroode/urlshorter/internal/model"
	"github.com/dtroode/urlshorter/internal/storage"
)

func TestStorage_GetURL(t *testing.T) {
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
			expectedError: storage.ErrNotFound,
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
			s := Storage{
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

func TestStorage_GetURLs(t *testing.T) {
	tests := map[string]struct {
		urlmap           URLMap
		shortKeys        []string
		expectedResponse []*model.URL
	}{
		"not all urls found": {
			urlmap: URLMap{
				"ydx": &model.URL{
					ShortKey:    "ydx",
					OriginalURL: "yandex.ru",
				},
				"ggl": &model.URL{
					ShortKey:    "ggl",
					OriginalURL: "google.com",
				},
			},
			shortKeys: []string{"ydx", "apl"},
			expectedResponse: []*model.URL{
				{
					ShortKey:    "ydx",
					OriginalURL: "yandex.ru",
				},
			},
		},
		"all urls found": {
			urlmap: URLMap{
				"ydx": &model.URL{
					ShortKey:    "ydx",
					OriginalURL: "yandex.ru",
				},
				"ggl": &model.URL{
					ShortKey:    "ggl",
					OriginalURL: "google.com",
				},
			},
			shortKeys: []string{"ydx", "ggl"},
			expectedResponse: []*model.URL{
				{
					ShortKey:    "ydx",
					OriginalURL: "yandex.ru",
				},
				{
					ShortKey:    "ggl",
					OriginalURL: "google.com",
				},
			},
		},
	}

	for tn, tt := range tests {
		t.Run(tn, func(t *testing.T) {
			t.Parallel()

			s := Storage{
				urlmap: tt.urlmap,
			}

			urls, err := s.GetURLs(context.Background(), tt.shortKeys)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResponse, urls)
		})
	}
}

func TestStorage_GetURLsByUserID(t *testing.T) {
	tests := map[string]struct {
		urlmap           URLMap
		userID           uuid.UUID
		expectedResponse []*model.URL
	}{
		"success": {
			urlmap: URLMap{
				"ydx": &model.URL{
					ShortKey:    "ydx",
					OriginalURL: "yandex.ru",
					UserID:      uuid.New(),
				},
				"ggl": &model.URL{
					ShortKey:    "ggl",
					OriginalURL: "google.com",
					UserID:      uuid.Max,
				},
			},
			userID: uuid.Max,
			expectedResponse: []*model.URL{
				{
					ShortKey:    "ggl",
					OriginalURL: "google.com",
					UserID:      uuid.Max,
				},
			},
		},
	}

	for tn, tt := range tests {
		t.Run(tn, func(t *testing.T) {
			t.Parallel()

			s := Storage{
				urlmap: tt.urlmap,
			}

			urls, err := s.GetURLsByUserID(context.Background(), tt.userID)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResponse, urls)
		})
	}
}

type dummyFile struct {
	*bytes.Buffer
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
			s := Storage{
				urlmap:  tt.urlmap,
				file:    &dummyFile{Buffer: buf},
				encoder: json.NewEncoder(buf),
			}

			url, err := s.SetURL(context.Background(), tt.url)
			require.NoError(t, err)

			assert.Equal(t, tt.url, (tt.urlmap)[tt.url.ShortKey])
			assert.Equal(t, tt.url, url)

			line, err := buf.ReadBytes('\n')
			require.NoError(t, err)

			writtenData := &model.URL{}
			err = json.Unmarshal(line, writtenData)
			require.NoError(t, err)

			assert.Equal(t, tt.url, writtenData)
		})
	}
}

func TestURL_SetURLs(t *testing.T) {
	buf := bytes.NewBuffer(nil)

	tests := map[string]struct {
		urlmap URLMap
		urls   []*model.URL
	}{
		"success": {
			urlmap: URLMap{
				"id1": &model.URL{
					ShortKey:    "abcd1",
					OriginalURL: "yandex.ru",
				},
			},
			urls: []*model.URL{
				{
					ShortKey:    "abcd2",
					OriginalURL: "yandex.ru",
				},
			},
		},
	}

	for tn, tt := range tests {
		t.Run(tn, func(t *testing.T) {
			s := Storage{
				urlmap:  tt.urlmap,
				file:    &dummyFile{Buffer: buf},
				encoder: json.NewEncoder(buf),
			}

			savedURLs, err := s.SetURLs(context.Background(), tt.urls)
			require.NoError(t, err)

			assert.Equal(t, tt.urls, savedURLs)

			for _, u := range tt.urls {
				assert.Equal(t, u, (tt.urlmap)[u.ShortKey])

				l := buf.Len()
				fmt.Println(l)

				line, err := buf.ReadBytes('\n')
				require.NoError(t, err)

				writtenData := &model.URL{}
				err = json.Unmarshal(line, writtenData)
				require.NoError(t, err)

				assert.Equal(t, u, writtenData)
			}
		})
	}
}

func TestURL_DeleteURLs(t *testing.T) {
	ids := []uuid.UUID{uuid.New(), uuid.New()}
	s := Storage{
		urlmap: URLMap{
			"ydx": &model.URL{
				ID:          ids[0],
				ShortKey:    "ydx",
				OriginalURL: "yandex.ru",
			},
		},
	}
	err := s.DeleteURLs(context.Background(), ids)
	require.NoError(t, err)
}

func TestURLMap_UnmarshalJSON(t *testing.T) {
	jsonData := `[{"short_key":"abc","original_url":"https://ya.ru"},{"short_key":"def","original_url":"https://google.com"}]`
	m := URLMap{}
	err := m.UnmarshalJSON([]byte(jsonData))
	require.NoError(t, err)
	assert.Len(t, m, 2)
	assert.Equal(t, "https://ya.ru", m["abc"].OriginalURL)
	assert.Equal(t, "https://google.com", m["def"].OriginalURL)
}

func TestStorage_Ping(t *testing.T) {
	s := &Storage{}
	err := s.Ping(context.Background())
	assert.NoError(t, err)
}

func TestStorage_NewStorage_And_Close(t *testing.T) {
	filename := "test_storage_file.json"
	defer func() { _ = os.Remove(filename) }()

	s, err := NewStorage(filename)
	require.NoError(t, err)
	require.NotNil(t, s)

	err = s.Close()
	assert.NoError(t, err)
}

func TestStorage_Count(t *testing.T) {
	userID1 := uuid.New()
	userID2 := uuid.New()
	s := Storage{
		urlmap: URLMap{
			"key1": {ID: uuid.New(), ShortKey: "countkey1", OriginalURL: "https://count1.com", UserID: userID1},
			"key2": {ID: uuid.New(), ShortKey: "countkey2", OriginalURL: "https://count2.com", UserID: userID1},
			"key3": {ID: uuid.New(), ShortKey: "countkey3", OriginalURL: "https://count3.com", UserID: userID2},
			"key4": {ID: uuid.New(), ShortKey: "countkey4", OriginalURL: "https://count4.com"},
		},
	}

	t.Run("count urls", func(t *testing.T) {
		count, err := s.CountURLs(context.Background())
		require.NoError(t, err)
		assert.Equal(t, int64(4), count)
	})

	t.Run("count users", func(t *testing.T) {
		count, err := s.CountUsers(context.Background())
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})
}
