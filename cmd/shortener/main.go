package main

import (
	"log"
	"net/http"

	"github.com/dtroode/urlshorter/config"
	internalLogger "github.com/dtroode/urlshorter/internal/logger"
	"github.com/dtroode/urlshorter/internal/router"
	"github.com/dtroode/urlshorter/internal/service"
	"github.com/dtroode/urlshorter/internal/storage"
)

func main() {
	config, err := config.ParseFlags()
	if err != nil {
		log.Fatal(err)
	}

	logger, err := internalLogger.NewLog(config.LogLevel)
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()

	urlStorage := storage.NewURL()

	urlService := service.NewURL(config.BaseURL, config.ShortURLLength, urlStorage)

	r := router.NewRouter()

	r.RegisterRoutes(urlService, logger)
	r.RegisterAPIRoutes(urlService, logger)

	logger.Infow("server started", "address", config.RunAddr)
	err = http.ListenAndServe(config.RunAddr, r)
	if err != nil {
		logger.Fatal(err)
	}
}
