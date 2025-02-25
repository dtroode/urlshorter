package main

import (
	"net/http"

	"github.com/dtroode/urlshorter/internal/app"
)

const ShortURLLength = 8
const BaseURL = "http://localhost:8080/"

func main() {

	mux := http.NewServeMux()

	serivce := app.NewService(BaseURL, ShortURLLength)

	mux.Handle("POST /{$}", NewCreateShorURLHandler(serivce))
	mux.Handle("GET /{id}", NewGetShortURLHandler(serivce))

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
