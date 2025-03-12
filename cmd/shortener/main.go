package main

import (
	"log"
	"net/http"

	"github.com/dtroode/urlshorter/config"
	"github.com/dtroode/urlshorter/internal/handler"
	"github.com/dtroode/urlshorter/internal/service"
	"github.com/dtroode/urlshorter/internal/storage"
	"github.com/go-chi/chi/v5"
)

func main() {
	config, err := config.ParseFlags()
	if err != nil {
		log.Fatal(err)
	}

	urlStorage := storage.NewURL()

	urlService := service.NewURL(config.BaseURL, config.ShortURLLength, urlStorage)

	h := handler.NewHandler(urlService)

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Post("/", h.CreateShortURL)
		r.Get("/{id}", h.GetShortURL)
	})

	err = http.ListenAndServe(config.RunAddr, r)
	if err != nil {
		log.Fatal(err)
	}
}
