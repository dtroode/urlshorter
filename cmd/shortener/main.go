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
	buildVersion string
	buildDate    string
	buildCommit  string
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
			logger.Error("failed to create database storage", "error", err)
			log.Fatal(err)
		}
		logger.Debug("using database storage")
	} else {
		urlStorage, err = inmemory.NewStorage(config.FileStoragePath)
		if err != nil {
			logger.Error("failed to create inmemory storage", "error", err, "file", config.FileStoragePath)
			log.Fatal(err)
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
		err = http.ListenAndServe(config.RunAddr, r)
		if err != nil {
			logger.Error("error running server", "error", err)
			log.Fatal(err)
		}

	}()

	logger.Info("server started", "address", config.RunAddr)
	logAppVersion()

	<-sigChan
	logger.Info("received interruption signal, exitting")
}

func logAppVersion() {
	var (
		bv = buildVersion
		bd = buildDate
		bc = buildCommit
	)

	tmpl := `
Build version: %s
Build date: %s
Build commit: %s
`

	if bv == "" {
		bv = "N/A"
	}
	if bd == "" {
		bd = "N/A"
	}
	if bc == "" {
		bc = "N/A"
	}

	fmt.Printf(tmpl, bv, bd, bc)
}
