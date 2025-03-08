package handler

import (
	"context"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Service interface {
	CreateShortURL(ctx context.Context, originalURL string) (*string, error)
	GetOriginalURL(ctx context.Context, id string) (*string, error)
}

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) CreateShortURL(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) GetShortURL(w http.ResponseWriter, r *http.Request) {
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
