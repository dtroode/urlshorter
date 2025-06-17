package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dtroode/urlshorter/internal/model"
	"github.com/dtroode/urlshorter/internal/storage"
	"github.com/dtroode/urlshorter/internal/storage/postgres"
)

const SKIP = true
const DSN = "postgres://postgres:postgres@localhost:5432/test?sslmode=disable"

func createURL(ctx context.Context, t *testing.T, db *pgx.Conn, url *model.URL) {
	query := `
	INSERT INTO urls (id, short_key, original_url, user_id, deleted_at) VALUES (@id, @shortKey, @originalURL, @userID, @deletedAt)`
	args := pgx.NamedArgs{
		"id":          url.ID,
		"shortKey":    url.ShortKey,
		"originalURL": url.OriginalURL,
		"userID":      url.UserID,
		"deletedAt":   url.DeletedAt,
	}

	_, err := db.Exec(ctx, query, args)
	require.NoError(t, err)
}

func getURLByOriginal(ctx context.Context, t *testing.T, db *pgx.Conn, originalURL string) *model.URL {
	var url model.URL
	query := `SELECT id, short_key, original_url, user_id, deleted_at FROM urls WHERE original_url = $1`
	err := db.QueryRow(ctx, query, originalURL).Scan(&url.ID, &url.ShortKey, &url.OriginalURL, &url.UserID, &url.DeletedAt)
	require.NoError(t, err)

	return &url
}

func getURLs(ctx context.Context, t *testing.T, db *pgx.Conn, ids []uuid.UUID) []*model.URL {
	query := `SELECT id, short_key, original_url, user_id, deleted_at FROM urls WHERE id = ANY($1)`
	rows, err := db.Query(ctx, query, ids)
	require.NoError(t, err)
	defer rows.Close()

	urls := make([]*model.URL, 0)

	for rows.Next() {
		var url model.URL
		err := rows.Scan(&url.ID, &url.ShortKey, &url.OriginalURL, &url.UserID, &url.DeletedAt)
		require.NoError(t, err)
		urls = append(urls, &url)
	}

	return urls
}

func truncateTable(ctx context.Context, t *testing.T, db *pgx.Conn, table string) {
	query := "TRUNCATE TABLE " + pgx.Identifier.Sanitize([]string{table}) + " CASCADE"
	_, err := db.Exec(ctx, query)
	require.NoError(t, err)
}

func TestStorage_GetURL(t *testing.T) {
	if SKIP {
		t.Skip("skip so that ci does not fail")
	}

	ctx := context.TODO()
	db, err := pgx.Connect(ctx, DSN)
	require.NoError(t, err)

	deletedAt := time.Now()

	userID := uuid.New()
	url := &model.URL{
		ID:          uuid.New(),
		ShortKey:    "aboba",
		OriginalURL: "http://yandex.ru",
		UserID:      userID,
		DeletedAt:   &deletedAt,
	}

	t.Run("success", func(t *testing.T) {
		t.Cleanup(func() {
			truncateTable(ctx, t, db, "urls")
		})

		createURL(ctx, t, db, url)

		s, err := postgres.NewStorage(DSN)
		require.NoError(t, err)

		gotURL, err := s.GetURL(ctx, url.ShortKey)
		require.NoError(t, err)

		assert.Equal(t, url.ID, gotURL.ID)
		assert.Equal(t, url.ShortKey, gotURL.ShortKey)
		assert.Equal(t, url.OriginalURL, gotURL.OriginalURL)
		assert.Equal(t, url.UserID, gotURL.UserID)
		if gotURL.DeletedAt != nil && url.DeletedAt != nil {
			assert.True(t, gotURL.DeletedAt.Equal(*url.DeletedAt))
		}
	})

	t.Run("does not exist", func(t *testing.T) {
		t.Cleanup(func() {
			truncateTable(ctx, t, db, "urls")
		})

		s, err := postgres.NewStorage(DSN)
		require.NoError(t, err)

		_, err = s.GetURL(ctx, url.ShortKey)
		require.Equal(t, storage.ErrNotFound, err)
	})
}

func TestStorage_GetURLs(t *testing.T) {
	if SKIP {
		t.Skip("skip so that ci does not fail")
	}

	ctx := context.TODO()
	db, err := pgx.Connect(ctx, DSN)
	require.NoError(t, err)

	deletedAt := time.Now()
	userID := uuid.New()
	urls := []*model.URL{
		{
			ID:          uuid.New(),
			ShortKey:    "ydx",
			OriginalURL: "http://yandex.ru",
			UserID:      userID,
			DeletedAt:   &deletedAt,
		},
		{
			ID:          uuid.New(),
			ShortKey:    "ggl",
			OriginalURL: "http://google.com",
			UserID:      userID,
		},
	}

	t.Run("all exist", func(t *testing.T) {
		t.Cleanup(func() {
			truncateTable(ctx, t, db, "urls")
		})

		for _, url := range urls {
			createURL(ctx, t, db, url)
		}

		s, err := postgres.NewStorage(DSN)
		require.NoError(t, err)

		gotURLs, err := s.GetURLs(ctx, []string{"ydx", "ggl"})
		require.NoError(t, err)
		for i, gotURL := range gotURLs {
			assert.Equal(t, urls[i].ID, gotURL.ID)
			assert.Equal(t, urls[i].ShortKey, gotURL.ShortKey)
			assert.Equal(t, urls[i].OriginalURL, gotURL.OriginalURL)
			assert.Equal(t, urls[i].UserID, gotURL.UserID)
			if gotURL.DeletedAt != nil && urls[i].DeletedAt != nil {
				assert.True(t, gotURL.DeletedAt.Equal(*urls[i].DeletedAt))
			}
		}
	})

	t.Run("one exists", func(t *testing.T) {
		t.Cleanup(func() {
			truncateTable(ctx, t, db, "urls")
		})

		createURL(ctx, t, db, urls[0])

		s, err := postgres.NewStorage(DSN)
		require.NoError(t, err)

		gotURLs, err := s.GetURLs(ctx, []string{"ydx", "appl"})
		require.NoError(t, err)

		require.Len(t, gotURLs, 1)

		assert.Equal(t, urls[0].ID, gotURLs[0].ID)
		assert.Equal(t, urls[0].ShortKey, gotURLs[0].ShortKey)
		assert.Equal(t, urls[0].OriginalURL, gotURLs[0].OriginalURL)
		assert.Equal(t, urls[0].UserID, gotURLs[0].UserID)
		if gotURLs[0].DeletedAt != nil && urls[0].DeletedAt != nil {
			assert.True(t, gotURLs[0].DeletedAt.Equal(*urls[0].DeletedAt))
		}
	})
}

func TestStorage_GetURLsByUserID(t *testing.T) {
	if SKIP {
		t.Skip("skip so that ci does not fail")
	}

	ctx := context.TODO()
	db, err := pgx.Connect(ctx, DSN)
	require.NoError(t, err)

	userID := uuid.New()
	urls := []*model.URL{
		model.NewURL("aboba", "http://yandex.ru", userID),
		model.NewURL("abcde", "http://google.com", userID),
	}

	t.Run("exist", func(t *testing.T) {
		t.Cleanup(func() {
			truncateTable(ctx, t, db, "urls")

			for _, u := range urls {
				createURL(ctx, t, db, u)
			}

			s, err := postgres.NewStorage(DSN)
			require.NoError(t, err)

			gotURLs, err := s.GetURLsByUserID(ctx, userID)
			require.NoError(t, err)

			for i, u := range urls {
				assert.Equal(t, u.ID, gotURLs[i].ID)
				assert.Equal(t, u.ShortKey, gotURLs[i].ShortKey)
				assert.Equal(t, u.OriginalURL, gotURLs[i].OriginalURL)
				assert.Equal(t, u.UserID, gotURLs[i].UserID)
				if gotURLs[i].DeletedAt != nil && u.DeletedAt != nil {
					assert.True(t, gotURLs[i].DeletedAt.Equal(*u.DeletedAt))
				}
			}
		})
	})

	t.Run("no urls", func(t *testing.T) {
		truncateTable(ctx, t, db, "urls")

		s, err := postgres.NewStorage(DSN)
		require.NoError(t, err)

		gotURLs, err := s.GetURLsByUserID(ctx, userID)
		require.NoError(t, err)
		require.Empty(t, gotURLs)
	})
}

func TestStorage_SetURL(t *testing.T) {
	if SKIP {
		t.Skip("skip so that ci does not fail")
	}

	ctx := context.TODO()
	db, err := pgx.Connect(ctx, DSN)
	require.NoError(t, err)

	userID := uuid.New()
	url := model.NewURL("aboba", "http://yandex.ru", userID)

	t.Run("new", func(t *testing.T) {
		t.Cleanup(func() {
			truncateTable(ctx, t, db, "urls")
		})

		s, err := postgres.NewStorage(DSN)
		require.NoError(t, err)

		url, err = s.SetURL(ctx, url)
		require.NoError(t, err)

		gotURL := getURLByOriginal(ctx, t, db, url.OriginalURL)

		assert.Equal(t, url.ID, gotURL.ID)
		assert.Equal(t, url.ShortKey, gotURL.ShortKey)
		assert.Equal(t, url.OriginalURL, gotURL.OriginalURL)
		assert.Equal(t, url.UserID, gotURL.UserID)
	})

	t.Run("existing", func(t *testing.T) {
		t.Cleanup(func() {
			truncateTable(ctx, t, db, "urls")
		})

		createURL(ctx, t, db, url)

		s, err := postgres.NewStorage(DSN)
		require.NoError(t, err)

		// same user
		newURLSameUser := model.NewURL("abcde", "http://yandex.ru", userID)
		require.NotEqual(t, url.ShortKey, newURLSameUser.ShortKey)

		url, err = s.SetURL(ctx, newURLSameUser)
		require.Equal(t, storage.ErrConflict, err)

		gotURL := getURLByOriginal(ctx, t, db, url.OriginalURL)

		assert.Equal(t, url.ID, gotURL.ID)
		assert.Equal(t, url.ShortKey, gotURL.ShortKey)
		assert.Equal(t, url.OriginalURL, gotURL.OriginalURL)
		assert.Equal(t, url.UserID, gotURL.UserID)

		// different user
		newURLDifferentUser := model.NewURL("abubl", "http://yandex.ru", uuid.New())
		require.NotEqual(t, url.ShortKey, newURLDifferentUser.ShortKey)

		url, err = s.SetURL(ctx, newURLDifferentUser)
		require.Equal(t, storage.ErrConflict, err)

		gotURL = getURLByOriginal(ctx, t, db, url.OriginalURL)

		assert.Equal(t, url.ID, gotURL.ID)
		assert.Equal(t, url.ShortKey, gotURL.ShortKey)
		assert.Equal(t, url.OriginalURL, gotURL.OriginalURL)
		assert.Equal(t, url.UserID, gotURL.UserID)
	})
}

func TestStorage_SetURLs(t *testing.T) {
	if SKIP {
		t.Skip("skip so that ci does not fail")
	}

	ctx := context.TODO()
	db, err := pgx.Connect(ctx, DSN)
	require.NoError(t, err)

	urls := []*model.URL{
		model.NewURL("aboba", "http://yandex.ru", uuid.New()),
		model.NewURL("abcde", "http://google.com", uuid.New()),
	}

	t.Run("new", func(t *testing.T) {
		t.Cleanup(func() {
			truncateTable(ctx, t, db, "urls")
		})

		s, err := postgres.NewStorage(DSN)
		require.NoError(t, err)

		savedURLs, err := s.SetURLs(ctx, urls)
		require.NoError(t, err)

		for i, u := range urls {
			assert.Equal(t, u.ID, savedURLs[i].ID)
			assert.Equal(t, u.ShortKey, savedURLs[i].ShortKey)
			assert.Equal(t, u.OriginalURL, savedURLs[i].OriginalURL)
			assert.Equal(t, u.UserID, savedURLs[i].UserID)

			gotURL := getURLByOriginal(ctx, t, db, u.OriginalURL)

			assert.Equal(t, u.ID, gotURL.ID)
			assert.Equal(t, u.ShortKey, gotURL.ShortKey)
			assert.Equal(t, u.OriginalURL, gotURL.OriginalURL)
			assert.Equal(t, u.UserID, gotURL.UserID)
		}
	})

	t.Run("existing", func(t *testing.T) {
		t.Cleanup(func() {
			truncateTable(ctx, t, db, "urls")
		})

		preURL := model.NewURL("abubu", "http://yandex.ru", uuid.New())
		createURL(ctx, t, db, preURL)

		s, err := postgres.NewStorage(DSN)
		require.NoError(t, err)

		savedURLs, err := s.SetURLs(ctx, urls)
		require.NoError(t, err)
		require.Equal(t, preURL.OriginalURL, urls[0].OriginalURL)

		gotExistedURL := getURLByOriginal(ctx, t, db, urls[0].OriginalURL)

		assert.Equal(t, preURL.ID, gotExistedURL.ID)
		assert.Equal(t, preURL.ShortKey, gotExistedURL.ShortKey)
		assert.Equal(t, preURL.OriginalURL, gotExistedURL.OriginalURL)
		assert.Equal(t, preURL.UserID, gotExistedURL.UserID)

		assert.Equal(t, urls[0].OriginalURL, savedURLs[0].OriginalURL)

		gotURL := getURLByOriginal(ctx, t, db, urls[1].OriginalURL)

		assert.Equal(t, urls[1].ID, gotURL.ID)
		assert.Equal(t, urls[1].ShortKey, gotURL.ShortKey)
		assert.Equal(t, urls[1].OriginalURL, gotURL.OriginalURL)
		assert.Equal(t, urls[1].UserID, gotURL.UserID)

		assert.Equal(t, urls[1].ID, savedURLs[1].ID)
		assert.Equal(t, urls[1].ShortKey, savedURLs[1].ShortKey)
		assert.Equal(t, urls[1].OriginalURL, savedURLs[1].OriginalURL)
		assert.Equal(t, urls[1].UserID, savedURLs[1].UserID)
	})
}

func TestStorage_DeleteURLs(t *testing.T) {
	if SKIP {
		t.Skip("skip so that ci does not fail")
	}

	ctx := context.TODO()
	db, err := pgx.Connect(ctx, DSN)
	require.NoError(t, err)

	ids := []uuid.UUID{uuid.New(), uuid.New()}

	urls := []*model.URL{
		model.NewURL("aboba", "http://yandex.ru", ids[0]),
		model.NewURL("abcde", "http://google.com", ids[1]),
	}

	t.Run("success", func(t *testing.T) {
		t.Cleanup(func() {
			truncateTable(ctx, t, db, "urls")
		})

		for _, url := range urls {
			createURL(ctx, t, db, url)
		}

		gotURLs := getURLs(ctx, t, db, ids)

		for _, url := range gotURLs {
			require.True(t, url.DeletedAt.IsZero())
		}

		s, err := postgres.NewStorage(DSN)
		require.NoError(t, err)

		err = s.DeleteURLs(ctx, ids)
		require.NoError(t, err)

		gotURLs = getURLs(ctx, t, db, ids)

		for _, url := range gotURLs {
			require.False(t, url.DeletedAt.IsZero())
		}
	})
}
