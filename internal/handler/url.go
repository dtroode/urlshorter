package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/dtroode/urlshorter/internal/auth"
	"github.com/dtroode/urlshorter/internal/logger"
	"github.com/dtroode/urlshorter/internal/request"
	"github.com/dtroode/urlshorter/internal/response"
	"github.com/dtroode/urlshorter/internal/service"
)

type URLService interface {
	GetOriginalURL(ctx context.Context, id string) (string, error)
	GetUserURLs(ctx context.Context, userID uuid.UUID) ([]*response.GetUserURL, error)
	CreateShortURL(ctx context.Context, dto *service.CreateShortURLDTO) (string, error)
	CreateShortURLBatch(ctx context.Context, dto *service.CreateShortURLBatchDTO) ([]*response.CreateShortURLBatch, error)
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
		if errors.Is(err, service.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)

			return
		}
		h.logger.Error("service error", "error", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.Header().Set("location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *URL) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := auth.GetUserIDFromContext(ctx)
	if !ok {
		h.logger.Error("failed to get user id from context")
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	url := string(body)
	dto := service.NewCreateShortURLDTO(url, userID)
	shortURL, err := h.service.CreateShortURL(ctx, dto)
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

	w.Write([]byte(shortURL))
}

func (h *URL) CreateShortURLJSON(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := auth.GetUserIDFromContext(ctx)
	if !ok {
		h.logger.Error("failed to get user id from context")
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	request := request.CreateShortURL{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Info("failed to decode request")
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	dto := service.NewCreateShortURLDTO(request.URL, userID)
	shortURL, err := h.service.CreateShortURL(ctx, dto)
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
		URL: shortURL,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode response", "error", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}
}

func (h *URL) CreateShortURLBatch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := auth.GetUserIDFromContext(ctx)
	if !ok {
		h.logger.Error("failed to get user id from context")
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

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

	dto := service.NewCreateShortURLBatchDTO(request, userID)
	shortURLs, err := h.service.CreateShortURLBatch(ctx, dto)
	if err != nil {
		h.logger.Error("service error", "error", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(shortURLs); err != nil {
		h.logger.Error("failed to encode response", "error", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}
}

func (h *URL) GetUserURLs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := auth.GetUserIDFromContext(ctx)
	if !ok {
		h.logger.Error("failed to get user id from context")
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	userURLs, err := h.service.GetUserURLs(ctx, userID)
	if err != nil {
		if errors.Is(err, service.ErrNoContent) {
			w.WriteHeader(http.StatusNoContent)

			return
		}

		h.logger.Error("service error", "error", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(userURLs); err != nil {
		h.logger.Error("failed to encode response", "error", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}
}
