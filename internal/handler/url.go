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
	"github.com/dtroode/urlshorter/internal/service/dto"
)

// URLService defines the interface for URL shortening operations.
// It provides methods for creating, retrieving, and managing shortened URLs.
type URLService interface {
	// GetOriginalURL retrieves the original URL associated with a short key.
	// Returns the original URL string or an error if not found or deleted.
	GetOriginalURL(ctx context.Context, shortKey string) (string, error)

	// GetUserURLs retrieves all URLs created by a specific user.
	// Returns a slice of user URLs or an error if the operation fails.
	GetUserURLs(ctx context.Context, userID uuid.UUID) ([]*response.GetUserURL, error)

	// CreateShortURL creates a new shortened URL.
	// Returns the shortened URL string or an error if creation fails.
	CreateShortURL(ctx context.Context, dto *dto.CreateShortURL) (string, error)

	// CreateShortURLBatch creates multiple shortened URLs in a single operation.
	// Returns a slice of created URLs with their correlation IDs or an error if creation fails.
	CreateShortURLBatch(ctx context.Context, dto *dto.CreateShortURLBatch) ([]*response.CreateShortURLBatch, error)

	// DeleteURLs marks the specified URLs as deleted for the given user.
	// Returns an error if the deletion operation fails.
	DeleteURLs(ctx context.Context, dto *dto.DeleteURLs) error
}

// URL represents the URL shortening HTTP handler.
// It provides HTTP endpoints for URL shortening operations.
type URL struct {
	service URLService
	logger  *logger.Logger
}

// NewURL creates a new URL handler instance with the provided service and logger.
func NewURL(s URLService, l *logger.Logger) *URL {
	return &URL{
		service: s,
		logger:  l,
	}
}

// GetOriginalURL handles GET requests to retrieve the original URL from a short key.
// @Summary Get original URL by short key
// @Description Redirects to the original URL associated with the provided short key
// @Tags URLs
// @Accept json
// @Produce json
// @Param id path string true "Short URL identifier"
// @Success 307 {string} string "Temporary redirect to original URL"
// @Failure 400 {string} string "Bad request - missing short key"
// @Failure 404 {string} string "URL not found"
// @Failure 410 {string} string "URL has been deleted"
// @Failure 500 {string} string "Internal server error"
// @Router /{id} [get]
func (h *URL) GetOriginalURL(w http.ResponseWriter, r *http.Request) {
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
		if errors.Is(err, service.ErrGone) {
			w.WriteHeader(http.StatusGone)

			return
		}
		h.logger.Error("service error", "error", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.Header().Set("location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// CreateShortURL handles POST requests to create a shortened URL from plain text.
// @Summary Create short URL from plain text
// @Description Creates a shortened URL from the provided plain text URL
// @Tags URLs
// @Accept text/plain
// @Produce text/plain
// @Param url body string true "Original URL to shorten"
// @Success 201 {string} string "Shortened URL created"
// @Success 409 {string} string "URL already exists"
// @Failure 400 {string} string "Bad request - invalid URL"
// @Failure 401 {string} string "Unauthorized - invalid or missing authentication"
// @Failure 500 {string} string "Internal server error"
// @Router / [post]
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
	dto := dto.NewCreateShortURL(url, userID)
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

// CreateShortURLJSON handles POST requests to create a shortened URL from JSON.
// @Summary Create short URL from JSON
// @Description Creates a shortened URL from the provided JSON request
// @Tags URLs
// @Accept json
// @Produce json
// @Param request body request.CreateShortURL true "URL shortening request"
// @Success 201 {object} response.CreateShortURL "Shortened URL created"
// @Success 409 {object} response.CreateShortURL "URL already exists"
// @Failure 400 {string} string "Bad request - invalid JSON"
// @Failure 401 {string} string "Unauthorized - invalid or missing authentication"
// @Failure 500 {string} string "Internal server error"
// @Router /api/shorten [post]
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

	dto := dto.NewCreateShortURL(request.URL, userID)
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

// CreateShortURLBatch handles POST requests to create multiple shortened URLs in batch.
// @Summary Create multiple short URLs in batch
// @Description Creates multiple shortened URLs from the provided batch request
// @Tags URLs
// @Accept json
// @Produce json
// @Param request body []request.CreateShortURLBatch true "Batch URL shortening request"
// @Success 201 {array} response.CreateShortURLBatch "Shortened URLs created"
// @Failure 400 {string} string "Bad request - invalid JSON or empty batch"
// @Failure 401 {string} string "Unauthorized - invalid or missing authentication"
// @Failure 500 {string} string "Internal server error"
// @Router /api/shorten/batch [post]
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

	dto := dto.NewCreateShortURLBatch(request, userID)
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

// GetUserURLs handles GET requests to retrieve all URLs created by the authenticated user.
// @Summary Get user's URLs
// @Description Retrieves all URLs created by the authenticated user
// @Tags User
// @Accept json
// @Produce json
// @Success 200 {array} response.GetUserURL "User's URLs"
// @Success 204 {string} string "No URLs found"
// @Failure 401 {string} string "Unauthorized - invalid or missing authentication"
// @Failure 500 {string} string "Internal server error"
// @Router /api/user/urls [get]
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

// DeleteURLs handles DELETE requests to mark URLs as deleted for the authenticated user.
// @Summary Delete user's URLs
// @Description Marks the specified URLs as deleted for the authenticated user
// @Tags User
// @Accept json
// @Produce json
// @Param shortKeys body []string true "Array of short keys to delete"
// @Success 202 {string} string "URLs marked for deletion"
// @Failure 400 {string} string "Bad request - invalid JSON"
// @Failure 401 {string} string "Unauthorized - invalid or missing authentication"
// @Failure 500 {string} string "Internal server error"
// @Router /api/user/urls [delete]
func (h *URL) DeleteURLs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := auth.GetUserIDFromContext(ctx)
	if !ok {
		h.logger.Error("failed to get user id from context")
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	var shortKeys []string
	if err := json.NewDecoder(r.Body).Decode(&shortKeys); err != nil {
		h.logger.Info("failed to decode request")
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	dto := dto.NewDeleteURLs(shortKeys, userID)
	if err := h.service.DeleteURLs(ctx, dto); err != nil {
		h.logger.Error("failed to delete urls", "error", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusAccepted)
}
