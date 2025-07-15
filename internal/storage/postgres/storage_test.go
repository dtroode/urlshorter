package postgres_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/dtroode/urlshorter/internal/model"
	"github.com/dtroode/urlshorter/internal/storage/postgres"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const skipIntegrationTests = true

var dsn string

func TestMain(m *testing.M) {
	if skipIntegrationTests {
		log.Println("Skipping postgres storage integration tests as configured.")
		return
	}

	ctx := context.Background()
	c, err := runPostgresContainer(ctx)
	if err != nil {
		log.Fatal(err)
	}

	host, err := c.Host(ctx)
	if err != nil {
		log.Fatalf("failed to get container host: %v", err)
	}
	port, err := c.MappedPort(ctx, "5432")
	if err != nil {
		log.Fatalf("failed to get container port: %v", err)
	}
	dsn = fmt.Sprintf("postgres://postgres:password@%s:%s/urlshorter_test?sslmode=disable", host, port.Port())

	code := m.Run()

	if err := c.Terminate(ctx); err != nil {
		log.Fatalf("failed to terminate container: %s", err)
	}

	os.Exit(code)
}

func runPostgresContainer(ctx context.Context) (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "password",
			"POSTGRES_DB":       "urlshorter_test",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(5 * time.Minute),
	}
	return testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
}

func TestStorage(t *testing.T) {
	s, err := postgres.NewStorage(dsn)
	require.NoError(t, err)
	defer s.Close()

	ctx := context.Background()

	t.Run("ping", func(t *testing.T) {
		err := s.Ping(ctx)
		require.NoError(t, err)
	})

	t.Run("set_and_get_url", func(t *testing.T) {
		userID := uuid.New()
		url := &model.URL{
			ID:          uuid.New(),
			ShortKey:    "testkey",
			OriginalURL: "https://example.com",
			UserID:      userID,
		}

		savedURL, err := s.SetURL(ctx, url)
		require.NoError(t, err)
		require.Equal(t, url.ID, savedURL.ID)

		retrievedURL, err := s.GetURL(ctx, "testkey")
		require.NoError(t, err)
		require.Equal(t, url.ID, retrievedURL.ID)
		require.NotNil(t, retrievedURL.UserID)
		require.Equal(t, userID, retrievedURL.UserID)
	})

	t.Run("set_urls_and_get_urls", func(t *testing.T) {
		userID := uuid.New()
		urls := []*model.URL{
			{ID: uuid.New(), ShortKey: "key1", OriginalURL: "https://ex1.com", UserID: userID},
			{ID: uuid.New(), ShortKey: "key2", OriginalURL: "https://ex2.com", UserID: userID},
		}

		savedURLs, err := s.SetURLs(ctx, urls)
		require.NoError(t, err)
		require.Len(t, savedURLs, 2)

		shortKeys := []string{"key1", "key2"}
		retrievedURLs, err := s.GetURLs(ctx, shortKeys)
		require.NoError(t, err)
		require.Len(t, retrievedURLs, 2)
	})

	t.Run("get_urls_by_user_id", func(t *testing.T) {
		userID := uuid.New()
		urls := []*model.URL{
			{ID: uuid.New(), ShortKey: "userkey1", OriginalURL: "https://user1.com", UserID: userID},
			{ID: uuid.New(), ShortKey: "userkey2", OriginalURL: "https://user2.com", UserID: userID},
		}

		_, err := s.SetURLs(ctx, urls)
		require.NoError(t, err)

		userURLs, err := s.GetURLsByUserID(ctx, userID)
		require.NoError(t, err)
		require.Len(t, userURLs, 2)
	})

	t.Run("delete_urls", func(t *testing.T) {
		userID := uuid.New()
		url := &model.URL{
			ID:          uuid.New(),
			ShortKey:    "deletekey",
			OriginalURL: "https://delete.com",
			UserID:      userID,
		}

		savedURL, err := s.SetURL(ctx, url)
		require.NoError(t, err)

		err = s.DeleteURLs(ctx, []uuid.UUID{savedURL.ID})
		require.NoError(t, err)

		retrievedURL, err := s.GetURL(ctx, "deletekey")
		require.NoError(t, err)
		require.NotNil(t, retrievedURL.DeletedAt, "DeletedAt should not be nil after deletion")
	})

	t.Run("set_url_conflict", func(t *testing.T) {
		userID := uuid.New()
		url1 := &model.URL{
			ID:          uuid.New(),
			ShortKey:    "conflictkey",
			OriginalURL: "https://conflict.com",
			UserID:      userID,
		}
		_, err := s.SetURL(ctx, url1)
		require.NoError(t, err)

		url2 := &model.URL{
			ID:          uuid.New(),
			ShortKey:    "anotherkey",
			OriginalURL: "https://conflict.com", // Same original URL
			UserID:      userID,
		}
		savedURL, err := s.SetURL(ctx, url2)
		require.Error(t, err) // Expecting a conflict error
		require.Equal(t, url1.ID, savedURL.ID)
		require.Equal(t, "conflictkey", savedURL.ShortKey)
	})
}
