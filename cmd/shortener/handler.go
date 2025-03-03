package main

import (
	"context"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Service interface {
	CreateShortURL(ctx context.Context, originalURL string) (*string, error)
	GetOriginalURL(ctx context.Context, shortURL string) (*string, error)
}

type BaseHandler struct {
	service Service
}

type CreateShortURLHandler struct {
	*BaseHandler
}

func NewCreateShorURLHandler(service Service) *CreateShortURLHandler {
	return &CreateShortURLHandler{
		&BaseHandler{
			service: service,
		},
	}
}

func (h *CreateShortURLHandler) Handle() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		body, err := io.ReadAll(r.Body)
		r.Body.Close()
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
}

type GetShortURLHandler struct {
	*BaseHandler
}

func NewGetShortURLHandler(service Service) *GetShortURLHandler {
	return &GetShortURLHandler{
		&BaseHandler{
			service: service,
		},
	}
}

func (h *GetShortURLHandler) Handle() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		id := chi.URLParam(r, "id")
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
}
