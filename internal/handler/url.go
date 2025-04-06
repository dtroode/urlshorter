package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dtroode/urlshorter/internal/logger"
	"github.com/dtroode/urlshorter/internal/request"
	"github.com/dtroode/urlshorter/internal/response"
	"github.com/dtroode/urlshorter/internal/service"
)

type URLService interface {
	CreateShortURL(ctx context.Context, originalURL string) (*string, error)
	GetOriginalURL(ctx context.Context, id string) (*string, error)
	CreateShortURLBatch(ctx context.Context, urls []*request.CreateShortURLBatch) ([]*response.CreateShortURLBatch, error)
}

type URL struct {
	service URLService
	logger  *logger.Logger
}

func NewURL(s URLService, l *logger.Logger) *URL {
	return &URL{
		service: s,
		logger:  l,
	}
}

func (h *URL) GetShortURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := chi.URLParam(r, "id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	originalURL, err := h.service.GetOriginalURL(ctx, id)
	if err != nil {
		h.logger.Error("service error", "error", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.Header().Set("location", *originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *URL) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	url := string(body)
	shortURL, err := h.service.CreateShortURL(ctx, url)
	if err != nil && !errors.Is(err, service.ErrConflict) {
		h.logger.Error("service error", "error", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.Header().Set("content-type", "text/plain")

	if errors.Is(err, service.ErrConflict) {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}

	w.Write([]byte(*shortURL))
}

func (h *URL) CreateShortURLJSON(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := request.CreateShortURL{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Info("failed to decode request")
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	shortURL, err := h.service.CreateShortURL(ctx, request.URL)
	if err != nil && !errors.Is(err, service.ErrConflict) {
		h.logger.Error("service error", "error", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.Header().Set("content-type", "application/json")
	if errors.Is(err, service.ErrConflict) {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}

	response := response.CreateShortURL{
		URL: *shortURL,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return
	}
}

func (h *URL) CreateShortURLBatch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := []*request.CreateShortURLBatch{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Info("failed to decode request")
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	if len(request) == 0 {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	shortURLs, err := h.service.CreateShortURLBatch(ctx, request)
	if err != nil {
		h.logger.Error("service error", "error", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(shortURLs); err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return
	}
}
