package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dtroode/urlshorter/internal/logger"
	"github.com/dtroode/urlshorter/internal/request"
	"github.com/dtroode/urlshorter/internal/response"
)

//go:generate mockgen -destination=mocks/mock_store.go -package=mocks . Service
type Service interface {
	CreateShortURL(ctx context.Context, originalURL string) (*string, error)
	GetOriginalURL(ctx context.Context, id string) (*string, error)
}

type Handler struct {
	service Service
	logger  *logger.Logger
}

func NewHandler(s Service, l *logger.Logger) *Handler {
	return &Handler{
		service: s,
		logger:  l,
	}
}

func (h *Handler) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	url := string(body)
	shortURL, err := h.service.CreateShortURL(ctx, url)
	if err != nil {
		h.logger.Error("service error", "error", err)
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
		h.logger.Error("service error", "error", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.Header().Set("location", *originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *Handler) CreateShortURLJSON(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := request.CreateShortURL{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	shortURL, err := h.service.CreateShortURL(ctx, request.URL)
	if err != nil {
		h.logger.Error("service error", "error", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := response.CreateShortURL{
		URL: *shortURL,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return
	}
}
