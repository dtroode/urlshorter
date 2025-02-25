package main

import (
	"context"
	"io"
	"net/http"

	"github.com/dtroode/urlshorter/internal/app"
)

type BaseHandler struct {
	service *app.Service
}

type CreateShortURLHandler struct {
	*BaseHandler
}

func NewCreateShorURLHandler(service *app.Service) *CreateShortURLHandler {
	return &CreateShortURLHandler{
		&BaseHandler{
			service: service,
		},
	}
}

func (h *CreateShortURLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	url := string(body)
	shortURL, err := h.service.CreateShortURL(ctx, url)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusCreated)

	w.Write([]byte(*shortURL))
}

type GetShortURLHandler struct {
	*BaseHandler
}

func NewGetShortURLHandler(service *app.Service) *GetShortURLHandler {
	return &GetShortURLHandler{
		&BaseHandler{
			service: service,
		},
	}
}

func (h *GetShortURLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	id := r.PathValue("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	originalURL, err := h.service.GetOriginalURL(ctx, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.Header().Set("location", *originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
