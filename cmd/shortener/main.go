package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/dtroode/urlshorter/config"
	"github.com/dtroode/urlshorter/internal/auth"
	"github.com/dtroode/urlshorter/internal/logger"
	"github.com/dtroode/urlshorter/internal/router"
	"github.com/dtroode/urlshorter/internal/service"
	"github.com/dtroode/urlshorter/internal/storage"
	"github.com/dtroode/urlshorter/internal/storage/inmemory"
	"github.com/dtroode/urlshorter/internal/storage/postgres"
)

var (
	buildVersion = "N/A" // set by ldflags
	buildDate    = "N/A" // set by ldflags
	buildCommit  = "N/A" // set by ldflags
)

func main() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	config, err := config.Initialize()
	if err != nil {
		log.Fatal(err)
	}

	logger := logger.NewLog(config.LogLevel)

	var urlStorage storage.Storage
	dsn := config.DatabaseDSN
	if dsn != "" {
		urlStorage, err = postgres.NewStorage(dsn)
		if err != nil {
			logger.Fatal("failed to create database storage", "error", err)
		}
		logger.Debug("using database storage")
	} else {
		urlStorage, err = inmemory.NewStorage(config.FileStoragePath)
		if err != nil {
			logger.Fatal("failed to create inmemory storage", "error", err, "file", config.FileStoragePath)
		}
		logger.Debug("using inmemory storage")
	}
	defer urlStorage.Close()

	urlService := service.NewURL(config.BaseURL, config.ShortKeyLength, config.ConcurrencyLimit, config.QueueSize, urlStorage)
	healthService := service.NewHealth(urlStorage)

	jwt := auth.NewJWT(config.JWTSecretKey)

	r := router.NewRouter()
	r.RegisterProfiler()
	r.RegisterAPIRoutes(urlService, jwt, logger)
	r.RegisterHealthRoutes(healthService, logger)

	go func() {
		var err error

		if config.EnableHTTPS {
			err = http.ListenAndServeTLS(config.RunAddr, config.CertFileName, config.PrivateKeyFileName, r)
		} else {
			err = http.ListenAndServe(config.RunAddr, r)
		}
		if err != nil {
			logger.Fatal("error running server", "error", err)
		}
	}()

	logger.Info("server started", "address", config.RunAddr, "tls", config.EnableHTTPS)
	logAppVersion()

	<-sigChan
	logger.Info("received interruption signal, exitting")
}

func logAppVersion() {
	tmpl := `
Build version: %s
Build date: %s
Build commit: %s
`

	fmt.Printf(tmpl, buildVersion, buildDate, buildCommit)
}
