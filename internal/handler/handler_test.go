package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dtroode/urlshorter/internal/handler/mocks"
	"github.com/dtroode/urlshorter/internal/logger"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type failReader struct {
}

func (m *failReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("failed to read")
}

func TestHandler_CreateShortURL(t *testing.T) {
	dummyLogger := &logger.Logger{
		Logger: slog.New(slog.NewJSONHandler(io.Discard, nil)),
	}

	url := "http://yandex.ru/"
	responseURL := "http://localhost:8080/d8398Sj3"

	tests := map[string]struct {
		body             io.Reader
		url              string
		readBodyResponse int
		readBodyError    error
		serviceResponse  *string
		serviceError     error
		wantError        bool
		wantContentType  string
		wantStatusCode   int
		wantResponse     []byte
	}{
		"failed to read body": {
			body:           &failReader{},
			readBodyError:  errors.New("read body error"),
			wantError:      true,
			wantStatusCode: http.StatusBadRequest,
		},
		"service error": {
			body:             strings.NewReader(url),
			readBodyResponse: 0,
			serviceError:     errors.New("service error"),
			wantError:        true,
			wantStatusCode:   http.StatusInternalServerError,
		},
		"success": {
			body:             strings.NewReader(url),
			readBodyResponse: 0,
			serviceResponse:  &responseURL,
			wantStatusCode:   http.StatusCreated,
			wantContentType:  "text/plain",
			wantResponse:     []byte(responseURL),
		},
	}
	for tn, tt := range tests {
		t.Run(tn, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", tt.body)
			w := httptest.NewRecorder()

			service := mocks.NewService(t)
			service.On("CreateShortURL", r.Context(), url).Maybe().Return(tt.serviceResponse, tt.serviceError)

			h := NewHandler(service, dummyLogger)

			h.CreateShortURL(w, r)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.wantStatusCode, res.StatusCode)

			if !tt.wantError {
				resBody, err := io.ReadAll(res.Body)
				require.NoError(t, err)

				assert.Equal(t, tt.wantResponse, resBody)
				assert.Equal(t, tt.wantContentType, res.Header.Get("content-type"))
			}
		})
	}
}

func TestHandler_GetShortURL(t *testing.T) {
	dummyLogger := &logger.Logger{
		Logger: slog.New(slog.NewJSONHandler(io.Discard, nil)),
	}

	responseURL := "http://yandex.ru/"

	tests := map[string]struct {
		id              string
		serviceResponse *string
		serviceError    error
		wantError       bool
		wantStatusCode  int
		wantResponse    string
	}{
		"id is empty": {
			id:             "",
			wantError:      true,
			wantStatusCode: http.StatusBadRequest,
		},
		"service error": {
			id:             "d8398Sj3",
			serviceError:   errors.New("service error"),
			wantError:      true,
			wantStatusCode: http.StatusInternalServerError,
		},
		"success": {
			id:              "d8398Sj3",
			serviceResponse: &responseURL,
			wantStatusCode:  http.StatusTemporaryRedirect,
			wantResponse:    responseURL,
		},
	}

	for tn, tt := range tests {
		t.Run(tn, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/"+tt.id, nil)

			// add chi context to basic context and
			// url param to chi context for handler
			// to be able to read id url param
			// otherwise handler fails with empty id and bad request
			chiContext := chi.NewRouteContext()
			chiContext.URLParams.Add("id", tt.id)
			ctx := context.WithValue(r.Context(), chi.RouteCtxKey, chiContext)
			r = r.WithContext(ctx)

			w := httptest.NewRecorder()

			service := mocks.NewService(t)
			service.On("GetOriginalURL", ctx, tt.id).Maybe().Return(tt.serviceResponse, tt.serviceError)

			h := NewHandler(service, dummyLogger)

			h.GetShortURL(w, r)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.wantStatusCode, res.StatusCode)

			if !tt.wantError {
				assert.Equal(t, tt.wantResponse, res.Header.Get("location"))
			}
		})
	}
}

func TestHandler_CreateShortURLJSON(t *testing.T) {
	dummyLogger := &logger.Logger{
		Logger: slog.New(slog.NewJSONHandler(io.Discard, nil)),
	}

	url := "http://yandex.ru/"
	responseURL := "http://localhost:8080/d8398Sj3"

	tests := map[string]struct {
		body            string
		serviceResponse *string
		serviceError    error
		wantError       bool
		wantStatusCode  int
		wantContentType string
		wantResponse    string
	}{
		"failed to decode body": {
			body:           "wrong body",
			wantError:      true,
			wantStatusCode: http.StatusBadRequest,
		},
		"service error": {
			body:           fmt.Sprintf(`{"url": "%s"}`, url),
			serviceError:   errors.New("service error"),
			wantError:      true,
			wantStatusCode: http.StatusInternalServerError,
		},
		// "failed to encode body": {
		// 	body:            fmt.Sprintf(`{"url": "%s"}`, url),
		// 	serviceResponse: &responseURL,
		// },
		"success": {
			body:            fmt.Sprintf(`{"url": "%s"}`, url),
			serviceResponse: &responseURL,
			wantStatusCode:  http.StatusCreated,
			wantContentType: "application/json",
			wantResponse:    fmt.Sprintf(`{"result": "%s"}`, responseURL),
		},
	}

	for tn, tt := range tests {
		t.Run(tn, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			service := mocks.NewService(t)
			service.On("CreateShortURL", r.Context(), url).Maybe().Return(tt.serviceResponse, tt.serviceError)

			h := NewHandler(service, dummyLogger)

			h.CreateShortURLJSON(w, r)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.wantStatusCode, res.StatusCode)

			if !tt.wantError {
				resBody, err := io.ReadAll(res.Body)
				require.NoError(t, err)

				assert.JSONEq(t, tt.wantResponse, string(resBody))
				assert.Equal(t, tt.wantContentType, res.Header.Get("content-type"))
			}
		})
	}
}
