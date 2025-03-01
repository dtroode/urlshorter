package main

import (
	"net/http"

	"github.com/dtroode/urlshorter/internal/app"
	"github.com/go-chi/chi/v5"
)

const ShortURLLength = 8
const BaseURL = "http://localhost:8080/"

func main() {
	serivce := app.NewService(BaseURL, ShortURLLength)

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Post("/", NewCreateShorURLHandler(serivce).Handle())
		r.Get("/{id}", NewGetShortURLHandler(serivce).Handle())
	})

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		panic(err)
	}
}
