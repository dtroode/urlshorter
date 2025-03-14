package main

import (
	"log"
	"net/http"

	"github.com/dtroode/urlshorter/config"
	"github.com/dtroode/urlshorter/internal/handler"
	"github.com/dtroode/urlshorter/internal/logger"
	"github.com/dtroode/urlshorter/internal/middleware"
	"github.com/dtroode/urlshorter/internal/service"
	"github.com/dtroode/urlshorter/internal/storage"
	"github.com/go-chi/chi/v5"
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

	h := handler.NewHandler(urlService)

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Post("/", middleware.RequestLog(h.CreateShortURL))
		r.Get("/{id}", middleware.RequestLog(h.GetShortURL))
	})

	err = http.ListenAndServe(config.RunAddr, r)
	if err != nil {
		logger.Log.Fatal(err)
	}
	logger.Log.Infow("server started", "address", config.RunAddr)
}
