package config

import (
	"flag"
	"log"
	"os"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	RunAddr        string `env:"SERVER_ADDRESS"`
	BaseURL        string `env:"BASE_URL"`
	ShortURLLength int    `env:"SHORT_URL_LENGTH"`
}

func ParseFlags() *Config {
	config := new(Config)

	flag.StringVar(&config.RunAddr, "a", ":8080", "address for app to run in format `host:port` or `:port`")
	flag.StringVar(&config.BaseURL, "b", "http://localhost:8080", "base url which goes before short url id")
	flag.IntVar(&config.ShortURLLength, "l", 8, "short url id length")
	flag.Parse()

	err := env.Parse(config)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	return config
}
