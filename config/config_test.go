package config

import (
	"encoding/json"
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
		configFile *Config
		wantConfig *Config
		wantError  bool
	}{
		"default values": {
			args: []string{"cmd"},
			wantConfig: &Config{
				RunAddr:            ":8080",
				BaseURL:            "http://localhost:8080",
				ShortKeyLength:     8,
				LogLevel:           "INFO",
				FileStoragePath:    os.TempDir() + "/urls",
				DatabaseDSN:        "",
				JWTSecretKey:       "a-string-secret-at-least-256-bits-long",
				ConcurrencyLimit:   3,
				QueueSize:          0,
				EnableHTTPS:        false,
				CertFileName:       "",
				PrivateKeyFileName: "",
			},
		},
		"with command line flags": {
			args: []string{"cmd", "-a", ":9090", "-b", "https://example.com", "-u", "10", "-l", "DEBUG", "-f", "/tmp/test.json", "-d", "postgres://test", "-j", "custom-secret", "-cl", "5", "-q", "100", "-s", "-sc", "cert.pem", "-sp", "key.pem"},
			wantConfig: &Config{
				RunAddr:            ":9090",
				BaseURL:            "https://example.com",
				ShortKeyLength:     10,
				LogLevel:           "DEBUG",
				FileStoragePath:    "/tmp/test.json",
				DatabaseDSN:        "postgres://test",
				JWTSecretKey:       "custom-secret",
				ConcurrencyLimit:   5,
				QueueSize:          100,
				EnableHTTPS:        true,
				CertFileName:       "cert.pem",
				PrivateKeyFileName: "key.pem",
			},
		},
		"with environment variables": {
			envVars: map[string]string{
				"SERVER_ADDRESS":        ":9090",
				"BASE_URL":              "https://example.com",
				"SHORT_URL_LENGTH":      "10",
				"LOG_LEVEL":             "DEBUG",
				"FILE_STORAGE_PATH":     "/tmp/test.json",
				"DATABASE_DSN":          "postgres://test",
				"JWT_SECRET_KEY":        "custom-secret",
				"CONCURRENCY_LIMIT":     "5",
				"QUEUE_SIZE":            "100",
				"ENABLE_HTTPS":          "true",
				"CERT_FILE_NAME":        "cert.pem",
				"PRIVATE_KEY_FILE_NAME": "key.pem",
			},
			args: []string{"cmd"},
			wantConfig: &Config{
				RunAddr:            ":9090",
				BaseURL:            "https://example.com",
				ShortKeyLength:     10,
				LogLevel:           "DEBUG",
				FileStoragePath:    "/tmp/test.json",
				DatabaseDSN:        "postgres://test",
				JWTSecretKey:       "custom-secret",
				ConcurrencyLimit:   5,
				QueueSize:          100,
				EnableHTTPS:        true,
				CertFileName:       "cert.pem",
				PrivateKeyFileName: "key.pem",
			},
		},
		"environment variables override flags": {
			envVars: map[string]string{
				"SERVER_ADDRESS": ":9090",
				"BASE_URL":       "https://example.com",
			},
			args: []string{"cmd", "-a", ":8080", "-b", "http://localhost:8080"},
			wantConfig: &Config{
				RunAddr:            ":9090",
				BaseURL:            "https://example.com",
				ShortKeyLength:     8,
				LogLevel:           "INFO",
				FileStoragePath:    os.TempDir() + "/urls",
				DatabaseDSN:        "",
				JWTSecretKey:       "a-string-secret-at-least-256-bits-long",
				ConcurrencyLimit:   3,
				QueueSize:          0,
				EnableHTTPS:        false,
				CertFileName:       "",
				PrivateKeyFileName: "",
			},
		},
		"with config file": {
			configFile: &Config{
				RunAddr:            ":7070",
				BaseURL:            "http://config-file.com",
				ShortKeyLength:     12,
				LogLevel:           "WARN",
				FileStoragePath:    "/config/storage.json",
				DatabaseDSN:        "postgres://config",
				JWTSecretKey:       "config-secret",
				ConcurrencyLimit:   7,
				QueueSize:          200,
				EnableHTTPS:        true,
				CertFileName:       "config-cert.pem",
				PrivateKeyFileName: "config-key.pem",
			},
			args: []string{"cmd", "-c", "test-config.json"},
			wantConfig: &Config{
				RunAddr:            ":7070",
				BaseURL:            "http://config-file.com",
				ShortKeyLength:     12,
				LogLevel:           "WARN",
				FileStoragePath:    "/config/storage.json",
				DatabaseDSN:        "postgres://config",
				JWTSecretKey:       "config-secret",
				ConcurrencyLimit:   7,
				QueueSize:          200,
				EnableHTTPS:        true,
				CertFileName:       "config-cert.pem",
				PrivateKeyFileName: "config-key.pem",
			},
		},
		"flags override config file": {
			configFile: &Config{
				RunAddr:            ":7070",
				BaseURL:            "http://config-file.com",
				ShortKeyLength:     12,
				LogLevel:           "WARN",
				FileStoragePath:    "/config/storage.json",
				DatabaseDSN:        "postgres://config",
				JWTSecretKey:       "config-secret",
				ConcurrencyLimit:   7,
				QueueSize:          200,
				EnableHTTPS:        true,
				CertFileName:       "config-cert.pem",
				PrivateKeyFileName: "config-key.pem",
			},
			args: []string{"cmd", "-c", "test-config.json", "-a", ":9090", "-b", "https://flags-override.com"},
			wantConfig: &Config{
				RunAddr:            ":9090",
				BaseURL:            "https://flags-override.com",
				ShortKeyLength:     12,
				LogLevel:           "WARN",
				FileStoragePath:    "/config/storage.json",
				DatabaseDSN:        "postgres://config",
				JWTSecretKey:       "config-secret",
				ConcurrencyLimit:   7,
				QueueSize:          200,
				EnableHTTPS:        true,
				CertFileName:       "config-cert.pem",
				PrivateKeyFileName: "config-key.pem",
			},
		},
		"environment variables override config file and flags": {
			configFile: &Config{
				RunAddr:            ":7070",
				BaseURL:            "http://config-file.com",
				ShortKeyLength:     12,
				LogLevel:           "WARN",
				FileStoragePath:    "/config/storage.json",
				DatabaseDSN:        "postgres://config",
				JWTSecretKey:       "config-secret",
				ConcurrencyLimit:   7,
				QueueSize:          200,
				EnableHTTPS:        true,
				CertFileName:       "config-cert.pem",
				PrivateKeyFileName: "config-key.pem",
			},
			envVars: map[string]string{
				"SERVER_ADDRESS": ":9090",
				"BASE_URL":       "https://env-override.com",
				"LOG_LEVEL":      "DEBUG",
			},
			args: []string{"cmd", "-c", "test-config.json", "-a", ":8080", "-b", "https://flags-override.com"},
			wantConfig: &Config{
				RunAddr:            ":9090",
				BaseURL:            "https://env-override.com",
				ShortKeyLength:     12,
				LogLevel:           "DEBUG",
				FileStoragePath:    "/config/storage.json",
				DatabaseDSN:        "postgres://config",
				JWTSecretKey:       "config-secret",
				ConcurrencyLimit:   7,
				QueueSize:          200,
				EnableHTTPS:        true,
				CertFileName:       "config-cert.pem",
				PrivateKeyFileName: "config-key.pem",
			},
		},
		"config file from environment variable": {
			configFile: &Config{
				RunAddr:            ":7070",
				BaseURL:            "http://env-config.com",
				ShortKeyLength:     12,
				LogLevel:           "WARN",
				FileStoragePath:    "/env/storage.json",
				DatabaseDSN:        "postgres://env-config",
				JWTSecretKey:       "env-config-secret",
				ConcurrencyLimit:   7,
				QueueSize:          200,
				EnableHTTPS:        true,
				CertFileName:       "env-cert.pem",
				PrivateKeyFileName: "env-key.pem",
			},
			envVars: map[string]string{
				"CONFIG": "test-config.json",
			},
			args: []string{"cmd"},
			wantConfig: &Config{
				RunAddr:            ":7070",
				BaseURL:            "http://env-config.com",
				ShortKeyLength:     12,
				LogLevel:           "WARN",
				FileStoragePath:    "/env/storage.json",
				DatabaseDSN:        "postgres://env-config",
				JWTSecretKey:       "env-config-secret",
				ConcurrencyLimit:   7,
				QueueSize:          200,
				EnableHTTPS:        true,
				CertFileName:       "env-cert.pem",
				PrivateKeyFileName: "env-key.pem",
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

			// Create config file if needed
			if tt.configFile != nil {
				configFileName := "test-config.json"
				// Check if config file is specified in args or env
				for i, arg := range tt.args {
					if arg == "-c" && i+1 < len(tt.args) {
						configFileName = tt.args[i+1]
						break
					}
				}
				if envConfigFile, set := tt.envVars["CONFIG"]; set {
					configFileName = envConfigFile
				}

				configData, err := json.Marshal(tt.configFile)
				require.NoError(t, err)

				err = os.WriteFile(configFileName, configData, 0644)
				require.NoError(t, err)
				defer os.Remove(configFileName)
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
		"invalid enable https": {
			envVars: map[string]string{
				"ENABLE_HTTPS": "invalid",
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

func TestInitialize_InvalidConfigFile(t *testing.T) {
	tests := map[string]struct {
		args        []string
		envVars     map[string]string
		filePath    string
		fileContent string
		wantError   bool
	}{
		"non-existent config file": {
			args:      []string{"cmd", "-c", "non-existent.json"},
			wantError: true,
		},
		"invalid json in config file": {
			args:        []string{"cmd", "-c", "invalid-config.json"},
			filePath:    "invalid-config.json",
			fileContent: `{"server_address": "invalid json`,
			wantError:   true,
		},
		"config file from environment variable": {
			args:        []string{"cmd"},
			envVars:     map[string]string{"CONFIG": "env-config.json"},
			filePath:    "env-config.json",
			fileContent: `{"server_address": "invalid json`,
			wantError:   true,
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

			// Create invalid config file if needed
			if tt.filePath != "" {
				err := os.WriteFile(tt.filePath, []byte(tt.fileContent), 0644)
				require.NoError(t, err)
				defer os.Remove(tt.filePath)
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

func TestSetDefaults(t *testing.T) {
	config := &Config{}
	config.setDefaults()

	expected := &Config{
		RunAddr:            ":8080",
		BaseURL:            "http://localhost:8080",
		ShortKeyLength:     8,
		LogLevel:           "INFO",
		FileStoragePath:    os.TempDir() + "/urls",
		DatabaseDSN:        "",
		JWTSecretKey:       "a-string-secret-at-least-256-bits-long",
		ConcurrencyLimit:   3,
		QueueSize:          0,
		EnableHTTPS:        false,
		CertFileName:       "",
		PrivateKeyFileName: "",
	}

	assert.Equal(t, expected, config)
}
