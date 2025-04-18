package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	RunAddr         string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	ShortKeyLength  int    `env:"SHORT_URL_LENGTH"`
	LogLevel        string `env:"LOG_LEVEL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
}

func Initialize() (*Config, error) {
	config := new(Config)

	flag.StringVar(&config.RunAddr, "a", ":8080", "address for app to run in format `host:port` or `:port`")
	flag.StringVar(&config.BaseURL, "b", "http://localhost:8080", "base url which goes before short url id")
	flag.IntVar(&config.ShortKeyLength, "u", 8, "short key length")
	flag.StringVar(&config.LogLevel, "l", "INFO", "log level")
	flag.StringVar(&config.FileStoragePath, "f", fmt.Sprintf(os.TempDir(), "urls"), "path to file where to store urls")
	flag.StringVar(&config.DatabaseDSN, "d", "", "string for connecting to postgres")
	flag.Parse()

	err := env.Parse(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
