package main

import (
	"net/http"

	"github.com/dtroode/urlshorter/cmd/shortener/config"
	"github.com/dtroode/urlshorter/internal/app"
	"github.com/go-chi/chi/v5"
)

func main() {
	config := config.ParseFlags()

	serivce := app.NewService(config.BaseURL, config.ShortURLLength)

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Post("/", NewCreateShorURLHandler(serivce).Handle())
		r.Get("/{id}", NewGetShortURLHandler(serivce).Handle())
	})

	err := http.ListenAndServe(config.RunAddr, r)
	if err != nil {
		panic(err)
	}
}
