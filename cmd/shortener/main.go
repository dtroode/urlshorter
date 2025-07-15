package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt)
	defer stop()

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
	defer func() {
		if err := urlStorage.Close(); err != nil {
			logger.Error("failed to close storage", "error", err)
		}
	}()

	urlService := service.NewURL(config.BaseURL, config.ShortKeyLength, config.ConcurrencyLimit, config.QueueSize, urlStorage)
	healthService := service.NewHealth(urlStorage)

	jwt := auth.NewJWT(config.JWTSecretKey)

	r := router.NewRouter()
	r.RegisterProfiler()
	r.RegisterAPIRoutes(urlService, jwt, logger)
	r.RegisterHealthRoutes(healthService, logger)

	server := &http.Server{
		Addr:         config.RunAddr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		var err error

		if config.EnableHTTPS {
			err = server.ListenAndServeTLS(config.CertFileName, config.PrivateKeyFileName)
		} else {
			err = server.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			logger.Fatal("error running server", "error", err)
		}
	}()

	logger.Info("server started", "address", config.RunAddr, "tls", config.EnableHTTPS)
	logAppVersion()

	<-ctx.Done()
	logger.Info("received interruption signal, shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("error during server shutdown", "error", err)
	}
}

func logAppVersion() {
	tmpl := `
Build version: %s
Build date: %s
Build commit: %s
`

	fmt.Printf(tmpl, buildVersion, buildDate, buildCommit)
}
