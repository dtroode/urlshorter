package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/dtroode/urlshorter/config"
	"github.com/dtroode/urlshorter/internal/logger"
	"github.com/dtroode/urlshorter/internal/router"
	"github.com/dtroode/urlshorter/internal/service"
	"github.com/dtroode/urlshorter/internal/storage"
)

func main() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	config, err := config.Initialize()
	if err != nil {
		log.Fatal(err)
	}

	logger := logger.NewLog(config.LogLevel)

	healthService := &service.Health{}

	var urlStorage storage.Storage
	dsn := config.DatabaseDSN
	if dsn != "" {
		databaseStorage, err := storage.NewDatabase(dsn)
		if err != nil {
			logger.Error("failed to create database storage", "error", err)
			os.Exit(1)
		}
		urlStorage = databaseStorage
		logger.Info("using database storage")

		healthService.DB = databaseStorage
	} else {
		urlStorage, err = storage.NewInMemory(config.FileStoragePath)
		if err != nil {
			logger.Error("failed to create inmemory storage", "error", err)
			os.Exit(1)
		}

		logger.Info("using inmemory storage")
	}
	defer urlStorage.Close()

	urlService := service.NewURL(config.BaseURL, config.ShortKeyLength, urlStorage)

	r := router.NewRouter()
	r.RegisterAPIRoutes(urlService, logger)
	r.RegisterHealthRoutes(healthService, logger)

	go func() {
		logger.Info("server started", "address", config.RunAddr)
		err = http.ListenAndServe(config.RunAddr, r)
		if err != nil {
			logger.Error("error running server", "error", err)
			os.Exit(1)
		}
	}()

	<-sigChan
	logger.Info("received interruption signal, exitting")
}
