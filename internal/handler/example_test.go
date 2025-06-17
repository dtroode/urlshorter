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

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/dtroode/urlshorter/internal/auth"
	"github.com/dtroode/urlshorter/internal/handler/mocks"
	"github.com/dtroode/urlshorter/internal/logger"
	"github.com/dtroode/urlshorter/internal/request"
)

// ExampleURL_GetOriginalURL demonstrates how to handle a GET request to retrieve the original URL.
func ExampleURL_GetOriginalURL() {
	// Create a mock service
	service := mocks.NewURLService(nil)

	// Create a logger
	logger := &logger.Logger{}

	// Create the handler
	handler := NewURL(service, logger)

	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "/abc123", nil)

	// Add chi context with URL parameter
	chiContext := chi.NewRouteContext()
	chiContext.URLParams.Add("id", "abc123")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, chiContext)
	req = req.WithContext(ctx)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Call the handler
	handler.GetOriginalURL(w, req)

	// Check the response
	resp := w.Result()
	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Location: %s\n", resp.Header.Get("location"))

	// Output:
	// Status: 307
	// Location: https://example.com/very-long-url-path
}

// ExampleURL_CreateShortURL demonstrates how to handle a POST request to create a shortened URL from plain text.
func ExampleURL_CreateShortURL() {
	// Create a mock service
	service := mocks.NewURLService(nil)

	// Create a logger
	logger := &logger.Logger{}

	// Create the handler
	handler := NewURL(service, logger)

	// Create a test request with plain text body
	body := strings.NewReader("https://example.com/very-long-url-path")
	req := httptest.NewRequest(http.MethodPost, "/", body)

	// Set user ID in context
	userID := uuid.New()
	ctx := auth.SetUserIDToContext(req.Context(), userID)
	req = req.WithContext(ctx)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Call the handler
	handler.CreateShortURL(w, req)

	// Check the response
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
	// Create a mock service
	service := mocks.NewURLService(nil)

	// Create a logger
	logger := &logger.Logger{}

	// Create the handler
	handler := NewURL(service, logger)

	// Create request body
	requestBody := request.CreateShortURL{
		URL: "https://example.com/very-long-url-path",
	}

	bodyBytes, _ := json.Marshal(requestBody)
	body := bytes.NewReader(bodyBytes)

	// Create a test request
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", body)
	req.Header.Set("Content-Type", "application/json")

	// Set user ID in context
	userID := uuid.New()
	ctx := auth.SetUserIDToContext(req.Context(), userID)
	req = req.WithContext(ctx)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Call the handler
	handler.CreateShortURLJSON(w, req)

	// Check the response
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
	// Create a mock service
	service := mocks.NewURLService(nil)

	// Create a logger
	logger := &logger.Logger{}

	// Create the handler
	handler := NewURL(service, logger)

	// Create request body
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

	// Create a test request
	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", body)
	req.Header.Set("Content-Type", "application/json")

	// Set user ID in context
	userID := uuid.New()
	ctx := auth.SetUserIDToContext(req.Context(), userID)
	req = req.WithContext(ctx)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Call the handler
	handler.CreateShortURLBatch(w, req)

	// Check the response
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
	// Create a mock service
	service := mocks.NewURLService(nil)

	// Create a logger
	logger := &logger.Logger{}

	// Create the handler
	handler := NewURL(service, logger)

	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)

	// Set user ID in context
	userID := uuid.New()
	ctx := auth.SetUserIDToContext(req.Context(), userID)
	req = req.WithContext(ctx)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Call the handler
	handler.GetUserURLs(w, req)

	// Check the response
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
	// Create a mock service
	service := mocks.NewURLService(nil)

	// Create a logger
	logger := &logger.Logger{}

	// Create the handler
	handler := NewURL(service, logger)

	// Create request body
	shortKeys := []string{"abc123", "def456"}
	bodyBytes, _ := json.Marshal(shortKeys)
	body := bytes.NewReader(bodyBytes)

	// Create a test request
	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", body)
	req.Header.Set("Content-Type", "application/json")

	// Set user ID in context
	userID := uuid.New()
	ctx := auth.SetUserIDToContext(req.Context(), userID)
	req = req.WithContext(ctx)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Call the handler
	handler.DeleteURLs(w, req)

	// Check the response
	resp := w.Result()

	fmt.Printf("Status: %d\n", resp.StatusCode)

	// Output:
	// Status: 202
}
