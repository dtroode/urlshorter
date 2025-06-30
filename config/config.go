package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env/v6"
)

// Config contains application configuration parameters.
type Config struct {
	RunAddr          string `env:"SERVER_ADDRESS"`
	BaseURL          string `env:"BASE_URL"`
	ShortKeyLength   int    `env:"SHORT_URL_LENGTH"`
	LogLevel         string `env:"LOG_LEVEL"`
	FileStoragePath  string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN      string `env:"DATABASE_DSN"`
	JWTSecretKey     string `env:"JWT_SECRET_KEY"`
	ConcurrencyLimit int    `env:"CONCURRENCY_LIMIT"`
	QueueSize        int    `env:"QUEUE_SIZE"`
}

// Initialize creates and initializes application configuration.
func Initialize() (*Config, error) {
	config := new(Config)

	flag.StringVar(&config.RunAddr, "a", ":8080", "address for app to run in format `host:port` or `:port`")
	flag.StringVar(&config.BaseURL, "b", "http://localhost:8080", "base url which goes before short url id")
	flag.IntVar(&config.ShortKeyLength, "u", 8, "short key length")
	flag.StringVar(&config.LogLevel, "l", "INFO", "log level")
	flag.StringVar(&config.FileStoragePath, "f", fmt.Sprintf("%s/urls", os.TempDir()), "path to file where to store urls")
	flag.StringVar(&config.DatabaseDSN, "d", "", "string for connecting to postgres")
	flag.StringVar(&config.JWTSecretKey, "j", "a-string-secret-at-least-256-bits-long", "key to sign jwt")
	flag.IntVar(&config.ConcurrencyLimit, "cl", 3, "number of workers in url service worker pool")
	flag.IntVar(&config.QueueSize, "q", 0, "worker pool jobs queue size, if not passed, will be set based on concurrency limit")

	flag.Parse()

	err := env.Parse(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
