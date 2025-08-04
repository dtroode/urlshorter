package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/caarlos0/env/v6"
)

// Config contains application configuration parameters.
type Config struct {
	HTTPRunAddr        string `env:"SERVER_ADDRESS" json:"server_address"`
	GRPCRunAddr        string `env:"GRPC_ADDRESS" json:"grpc_address"`
	BaseURL            string `env:"BASE_URL" json:"base_url"`
	LogLevel           string `env:"LOG_LEVEL" json:"log_level"`
	FileStoragePath    string `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	DatabaseDSN        string `env:"DATABASE_DSN" json:"database_dsn"`
	JWTSecretKey       string `env:"JWT_SECRET_KEY" json:"jwt_secret_key"`
	CertFileName       string `env:"CERT_FILE_NAME" json:"cert_file_name"`
	PrivateKeyFileName string `env:"PRIVATE_KEY_FILE_NAME" json:"private_key_file_name"`
	TrustedSubnet      string `env:"TRUSTED_SUBNET" json:"trusted_subnet"`
	ShortKeyLength     int    `env:"SHORT_URL_LENGTH" json:"short_url_length"`
	ConcurrencyLimit   int    `env:"CONCURRENCY_LIMIT" json:"concurrency_limit"`
	QueueSize          int    `env:"QUEUE_SIZE" json:"queue_size"`
	EnableHTTPS        bool   `env:"ENABLE_HTTPS" json:"enable_https"`
}

func (c *Config) setDefaults() {
	c.HTTPRunAddr = ":8080"
	c.GRPCRunAddr = ":8081"
	c.BaseURL = "http://localhost:8080"
	c.ShortKeyLength = 8
	c.LogLevel = "INFO"
	c.FileStoragePath = fmt.Sprintf("%s/urls", os.TempDir())
	c.DatabaseDSN = ""
	c.JWTSecretKey = "a-string-secret-at-least-256-bits-long"
	c.ConcurrencyLimit = 3
	c.QueueSize = 0
	c.EnableHTTPS = false
	c.CertFileName = ""
	c.PrivateKeyFileName = ""
	c.TrustedSubnet = ""
}

// Initialize creates and initializes application configuration.
// It loads configuration from multiple sources in the following order:
// 1. Default values
// 2. Configuration file (if specified)
// 3. Command line flags
// 4. Environment variables (highest priority)
//
// Returns a fully configured Config instance or an error if configuration loading fails.
func Initialize() (*Config, error) {
	config := &Config{}
	config.setDefaults()

	configFile := parseFlagsForConfigFile()

	if configFile != "" {
		if err := parseConfigFromFile(config, configFile); err != nil {
			return nil, fmt.Errorf("failed to parse config from file: %w", err)
		}
	}

	if err := parseCommandLineFlags(config); err != nil {
		return nil, fmt.Errorf("failed to parse command line flags: %w", err)
	}

	if err := env.Parse(config); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	return config, nil
}
func parseFlagsForConfigFile() string {
	var configFile string

	tempFlagSet := flag.NewFlagSet("temp", flag.ContinueOnError)
	tempFlagSet.StringVar(&configFile, "c", "", "config file name")

	_ = tempFlagSet.Parse(os.Args[1:])

	if envConfigFile, ok := os.LookupEnv("CONFIG"); ok {
		configFile = envConfigFile
	}

	return configFile
}

func parseCommandLineFlags(config *Config) error {
	flagSet := flag.NewFlagSet("app", flag.ContinueOnError)

	flagSet.String("c", "", "config file name")
	flagSet.StringVar(&config.HTTPRunAddr, "a", config.HTTPRunAddr, "address for http server to run in format `host:port` or `:port`")
	flagSet.StringVar(&config.GRPCRunAddr, "g", config.GRPCRunAddr, "address for grpc server to run in format `host:port` or `:port`")
	flagSet.StringVar(&config.BaseURL, "b", config.BaseURL, "base url which goes before short url id")
	flagSet.IntVar(&config.ShortKeyLength, "u", config.ShortKeyLength, "short key length")
	flagSet.StringVar(&config.LogLevel, "l", config.LogLevel, "log level")
	flagSet.StringVar(&config.FileStoragePath, "f", config.FileStoragePath, "path to file where to store urls")
	flagSet.StringVar(&config.DatabaseDSN, "d", config.DatabaseDSN, "string for connecting to postgres")
	flagSet.StringVar(&config.JWTSecretKey, "j", config.JWTSecretKey, "key to sign jwt")
	flagSet.IntVar(&config.ConcurrencyLimit, "cl", config.ConcurrencyLimit, "number of workers in url service worker pool")
	flagSet.IntVar(&config.QueueSize, "q", config.QueueSize, "worker pool jobs queue size, if not passed, will be set based on concurrency limit")
	flagSet.BoolVar(&config.EnableHTTPS, "s", config.EnableHTTPS, "should server serve https")
	flagSet.StringVar(&config.CertFileName, "sc", config.CertFileName, "cert file name")
	flagSet.StringVar(&config.PrivateKeyFileName, "sp", config.PrivateKeyFileName, "private key file name")
	flagSet.StringVar(&config.TrustedSubnet, "t", config.TrustedSubnet, "trusted subnet for internal api")

	return flagSet.Parse(os.Args[1:])
}

func parseConfigFromFile(config *Config, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file %s for reading: %w", filename, err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("failed to close file %s: %v", filename, err)
		}
	}()

	if err := json.NewDecoder(file).Decode(config); err != nil {
		return fmt.Errorf("failed to decode config file: %w", err)
	}

	return nil
}
