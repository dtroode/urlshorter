package config

import "flag"

type Config struct {
	RunAddr        string
	BaseURL        string
	ShortURLLength int
}

func ParseFlags() *Config {
	config := new(Config)

	flag.StringVar(&config.RunAddr, "a", ":8080", "address for app to run in format `host:port` or `:port`")
	flag.StringVar(&config.BaseURL, "b", "http://localhost:8080", "base url which goes before short url id")
	flag.IntVar(&config.ShortURLLength, "l", 8, "short url id length")
	flag.Parse()

	return config
}
