package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/dtroode/urlshorter/internal/auth"
	"github.com/dtroode/urlshorter/internal/handler/mocks"
	"github.com/dtroode/urlshorter/internal/logger"
	"github.com/dtroode/urlshorter/internal/request"
	"github.com/dtroode/urlshorter/internal/response"
)

// ExampleURL_GetOriginalURL demonstrates how to handle a GET request to retrieve the original URL.
func ExampleURL_GetOriginalURL() {
	service := mocks.NewURLService(&testing.T{})

	service.On("GetOriginalURL", mock.Anything, "abc123").Return("https://example.com/very-long-url-path", nil)

	logger := &logger.Logger{}

	handler := NewURL(service, logger)

	req := httptest.NewRequest(http.MethodGet, "/abc123", nil)

	chiContext := chi.NewRouteContext()
	chiContext.URLParams.Add("id", "abc123")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, chiContext)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handler.GetOriginalURL(w, req)

	resp := w.Result()
	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Location: %s\n", resp.Header.Get("location"))

	// Output:
	// Status: 307
	// Location: https://example.com/very-long-url-path
}

// ExampleURL_CreateShortURL demonstrates how to handle a POST request to create a shortened URL from plain text.
func ExampleURL_CreateShortURL() {
	service := mocks.NewURLService(&testing.T{})

	userID := uuid.New()
	service.On("CreateShortURL", mock.Anything, mock.AnythingOfType("*dto.CreateShortURL")).Return("https://shortener.example.com/abc123", nil)

	logger := &logger.Logger{}

	handler := NewURL(service, logger)

	body := strings.NewReader("https://example.com/very-long-url-path")
	req := httptest.NewRequest(http.MethodPost, "/", body)

	ctx := auth.SetUserIDToContext(req.Context(), userID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handler.CreateShortURL(w, req)

	resp := w.Result()
	bodyBytes, _ := io.ReadAll(resp.Body)

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Content-Type: %s\n", resp.Header.Get("content-type"))
	fmt.Printf("Response: %s\n", string(bodyBytes))

	// Output:
	// Status: 201
	// Content-Type: text/plain
	// Response: https://shortener.example.com/abc123
}

// ExampleURL_CreateShortURLJSON demonstrates how to handle a POST request to create a shortened URL from JSON.
func ExampleURL_CreateShortURLJSON() {
	service := mocks.NewURLService(&testing.T{})

	userID := uuid.New()
	service.On("CreateShortURL", mock.Anything, mock.AnythingOfType("*dto.CreateShortURL")).Return("https://shortener.example.com/abc123", nil)

	logger := &logger.Logger{}

	handler := NewURL(service, logger)

	requestBody := request.CreateShortURL{
		URL: "https://example.com/very-long-url-path",
	}

	bodyBytes, _ := json.Marshal(requestBody)
	body := bytes.NewReader(bodyBytes)

	req := httptest.NewRequest(http.MethodPost, "/api/shorten", body)
	req.Header.Set("Content-Type", "application/json")

	ctx := auth.SetUserIDToContext(req.Context(), userID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handler.CreateShortURLJSON(w, req)

	resp := w.Result()
	respBodyBytes, _ := io.ReadAll(resp.Body)

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Content-Type: %s\n", resp.Header.Get("content-type"))
	fmt.Printf("Response: %s\n", string(respBodyBytes))

	// Output:
	// Status: 201
	// Content-Type: application/json
	// Response: {"result":"https://shortener.example.com/abc123"}
}

// ExampleURL_CreateShortURLBatch demonstrates how to handle a POST request to create multiple shortened URLs in batch.
func ExampleURL_CreateShortURLBatch() {
	service := mocks.NewURLService(&testing.T{})

	userID := uuid.New()
	expectedResponse := []*response.CreateShortURLBatch{
		{
			CorrelationID: "req-123",
			ShortURL:      "https://shortener.example.com/abc123",
		},
		{
			CorrelationID: "req-456",
			ShortURL:      "https://shortener.example.com/def456",
		},
	}
	service.On("CreateShortURLBatch", mock.Anything, mock.AnythingOfType("*dto.CreateShortURLBatch")).Return(expectedResponse, nil)

	logger := &logger.Logger{}

	handler := NewURL(service, logger)

	requestBody := []*request.CreateShortURLBatch{
		{
			CorrelationID: "req-123",
			OriginalURL:   "https://example.com/very-long-url-path-1",
		},
		{
			CorrelationID: "req-456",
			OriginalURL:   "https://example.com/very-long-url-path-2",
		},
	}

	bodyBytes, _ := json.Marshal(requestBody)
	body := bytes.NewReader(bodyBytes)

	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", body)
	req.Header.Set("Content-Type", "application/json")

	ctx := auth.SetUserIDToContext(req.Context(), userID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handler.CreateShortURLBatch(w, req)

	resp := w.Result()
	respBodyBytes, _ := io.ReadAll(resp.Body)

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Content-Type: %s\n", resp.Header.Get("content-type"))
	fmt.Printf("Response: %s\n", string(respBodyBytes))

	// Output:
	// Status: 201
	// Content-Type: application/json
	// Response: [{"correlation_id":"req-123","short_url":"https://shortener.example.com/abc123"},{"correlation_id":"req-456","short_url":"https://shortener.example.com/def456"}]
}

// ExampleURL_GetUserURLs demonstrates how to handle a GET request to retrieve all URLs created by the user.
func ExampleURL_GetUserURLs() {
	service := mocks.NewURLService(&testing.T{})

	userID := uuid.New()
	expectedResponse := []*response.GetUserURL{
		{
			ShortURL:    "https://shortener.example.com/abc123",
			OriginalURL: "https://example.com/very-long-url-path-1",
		},
		{
			ShortURL:    "https://shortener.example.com/def456",
			OriginalURL: "https://example.com/very-long-url-path-2",
		},
	}
	service.On("GetUserURLs", mock.Anything, userID).Return(expectedResponse, nil)

	logger := &logger.Logger{}

	handler := NewURL(service, logger)

	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)

	ctx := auth.SetUserIDToContext(req.Context(), userID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handler.GetUserURLs(w, req)

	resp := w.Result()
	respBodyBytes, _ := io.ReadAll(resp.Body)

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Content-Type: %s\n", resp.Header.Get("content-type"))
	fmt.Printf("Response: %s\n", string(respBodyBytes))

	// Output:
	// Status: 200
	// Content-Type: application/json
	// Response: [{"short_url":"https://shortener.example.com/abc123","original_url":"https://example.com/very-long-url-path-1"},{"short_url":"https://shortener.example.com/def456","original_url":"https://example.com/very-long-url-path-2"}]
}

// ExampleURL_DeleteURLs demonstrates how to handle a DELETE request to mark URLs as deleted.
func ExampleURL_DeleteURLs() {
	service := mocks.NewURLService(&testing.T{})

	userID := uuid.New()
	service.On("DeleteURLs", mock.Anything, mock.AnythingOfType("*dto.DeleteURLs")).Return(nil)

	logger := &logger.Logger{}

	handler := NewURL(service, logger)

	shortKeys := []string{"abc123", "def456"}
	bodyBytes, _ := json.Marshal(shortKeys)
	body := bytes.NewReader(bodyBytes)

	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", body)
	req.Header.Set("Content-Type", "application/json")

	ctx := auth.SetUserIDToContext(req.Context(), userID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handler.DeleteURLs(w, req)

	resp := w.Result()

	fmt.Printf("Status: %d\n", resp.StatusCode)

	// Output:
	// Status: 202
}
