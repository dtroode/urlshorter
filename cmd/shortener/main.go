package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/dtroode/urlshorter/config"
	internalLogger "github.com/dtroode/urlshorter/internal/logger"
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

	logger, err := internalLogger.NewLog(config.LogLevel)
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()

	urlStorage, err := storage.NewURL(config.FileStoragePath)
	if err != nil {
		log.Fatal(err)
	}
	defer urlStorage.Flush()

	urlService := service.NewURL(config.BaseURL, config.ShortURLLength, urlStorage)

	r := router.NewRouter()

	r.RegisterRoutes(urlService, logger)
	r.RegisterAPIRoutes(urlService, logger)

	go func() {
		logger.Infow("server started", "address", config.RunAddr)
		err = http.ListenAndServe(config.RunAddr, r)
		if err != nil {
			logger.Fatal(err)
		}
	}()

	<-sigChan
	logger.Infow("received interruption signal, exitting")
}
