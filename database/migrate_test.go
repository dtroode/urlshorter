package database_test

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/dtroode/urlshorter/database"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const skipIntegrationTests = true

func TestMigrate(t *testing.T) {
	if skipIntegrationTests {
		t.Skip("Skipping migration tests as configured.")
		return
	}

	ctx := context.Background()
	c, err := runPostgresContainer(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Terminate(ctx)

	host, err := c.Host(ctx)
	require.NoError(t, err)
	port, err := c.MappedPort(ctx, "5432")
	require.NoError(t, err)
	dsn := fmt.Sprintf("postgres://postgres:password@%s:%s/urlshorter_test?sslmode=disable", host, port.Port())

	err = database.Migrate(ctx, dsn)
	require.NoError(t, err)
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
