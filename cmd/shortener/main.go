package main

import (
	"log"
	"net/http"

	"github.com/dtroode/urlshorter/config"
	"github.com/dtroode/urlshorter/internal/logger"
	"github.com/dtroode/urlshorter/internal/router"
	"github.com/dtroode/urlshorter/internal/service"
	"github.com/dtroode/urlshorter/internal/storage"
)

func main() {
	config, err := config.ParseFlags()
	if err != nil {
		log.Fatal(err)
	}

	logger.Initialize(config.LogLevel)
	defer logger.Log.Sync()

	urlStorage := storage.NewURL()

	urlService := service.NewURL(config.BaseURL, config.ShortURLLength, urlStorage)

	r := router.NewRouter()

	r.RegisterRoutes(urlService)
	r.RegisterAPIRoutes(urlService)

	logger.Log.Infow("server started", "address", config.RunAddr)
	err = http.ListenAndServe(config.RunAddr, r)
	if err != nil {
		logger.Log.Fatal(err)
	}
}
