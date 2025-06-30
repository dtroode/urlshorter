package config

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitialize(t *testing.T) {
	tests := map[string]struct {
		envVars    map[string]string
		args       []string
		wantConfig *Config
		wantError  bool
	}{
		"default values": {
			args: []string{"cmd"},
			wantConfig: &Config{
				RunAddr:          ":8080",
				BaseURL:          "http://localhost:8080",
				ShortKeyLength:   8,
				LogLevel:         "INFO",
				FileStoragePath:  os.TempDir() + "/urls",
				DatabaseDSN:      "",
				JWTSecretKey:     "a-string-secret-at-least-256-bits-long",
				ConcurrencyLimit: 3,
				QueueSize:        0,
			},
		},
		"with command line flags": {
			args: []string{"cmd", "-a", ":9090", "-b", "https://example.com", "-u", "10", "-l", "DEBUG", "-f", "/tmp/test.json", "-d", "postgres://test", "-j", "custom-secret", "-cl", "5", "-q", "100"},
			wantConfig: &Config{
				RunAddr:          ":9090",
				BaseURL:          "https://example.com",
				ShortKeyLength:   10,
				LogLevel:         "DEBUG",
				FileStoragePath:  "/tmp/test.json",
				DatabaseDSN:      "postgres://test",
				JWTSecretKey:     "custom-secret",
				ConcurrencyLimit: 5,
				QueueSize:        100,
			},
		},
		"with environment variables": {
			envVars: map[string]string{
				"SERVER_ADDRESS":    ":9090",
				"BASE_URL":          "https://example.com",
				"SHORT_URL_LENGTH":  "10",
				"LOG_LEVEL":         "DEBUG",
				"FILE_STORAGE_PATH": "/tmp/test.json",
				"DATABASE_DSN":      "postgres://test",
				"JWT_SECRET_KEY":    "custom-secret",
				"CONCURRENCY_LIMIT": "5",
				"QUEUE_SIZE":        "100",
			},
			args: []string{"cmd"},
			wantConfig: &Config{
				RunAddr:          ":9090",
				BaseURL:          "https://example.com",
				ShortKeyLength:   10,
				LogLevel:         "DEBUG",
				FileStoragePath:  "/tmp/test.json",
				DatabaseDSN:      "postgres://test",
				JWTSecretKey:     "custom-secret",
				ConcurrencyLimit: 5,
				QueueSize:        100,
			},
		},
		"environment variables override flags": {
			envVars: map[string]string{
				"SERVER_ADDRESS": ":9090",
				"BASE_URL":       "https://example.com",
			},
			args: []string{"cmd", "-a", ":8080", "-b", "http://localhost:8080"},
			wantConfig: &Config{
				RunAddr:          ":9090",
				BaseURL:          "https://example.com",
				ShortKeyLength:   8,
				LogLevel:         "INFO",
				FileStoragePath:  os.TempDir() + "/urls",
				DatabaseDSN:      "",
				JWTSecretKey:     "a-string-secret-at-least-256-bits-long",
				ConcurrencyLimit: 3,
				QueueSize:        0,
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Save original args and restore after test
			originalArgs := os.Args
			defer func() { os.Args = originalArgs }()

			// Set test args
			os.Args = tt.args

			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Reset flag.CommandLine to avoid conflicts between tests
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			// Run test
			config, err := Initialize()

			if tt.wantError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantConfig, config)

			// Clean up environment variables
			for key := range tt.envVars {
				os.Unsetenv(key)
			}
		})
	}
}

func TestInitialize_InvalidEnvVars(t *testing.T) {
	tests := map[string]struct {
		envVars map[string]string
	}{
		"invalid short key length": {
			envVars: map[string]string{
				"SHORT_URL_LENGTH": "invalid",
			},
		},
		"invalid concurrency limit": {
			envVars: map[string]string{
				"CONCURRENCY_LIMIT": "invalid",
			},
		},
		"invalid queue size": {
			envVars: map[string]string{
				"QUEUE_SIZE": "invalid",
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Save original args and restore after test
			originalArgs := os.Args
			defer func() { os.Args = originalArgs }()

			// Set test args
			os.Args = []string{"cmd"}

			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Reset flag.CommandLine to avoid conflicts between tests
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			// Run test
			_, err := Initialize()
			assert.Error(t, err)

			// Clean up environment variables
			for key := range tt.envVars {
				os.Unsetenv(key)
			}
		})
	}
}
