package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dtroode/urlshorter/internal/auth"
	"github.com/dtroode/urlshorter/internal/handler/mocks"
	"github.com/dtroode/urlshorter/internal/logger"
	"github.com/dtroode/urlshorter/internal/request"
	"github.com/dtroode/urlshorter/internal/response"
	"github.com/dtroode/urlshorter/internal/service"
	"github.com/dtroode/urlshorter/internal/service/dto"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type failReader struct {
}

func (m *failReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("failed to read")
}

func TestHandler_GetOriginalURL(t *testing.T) {
	dummyLogger := &logger.Logger{
		Logger: slog.New(slog.NewJSONHandler(io.Discard, nil)),
	}

	responseURL := "http://yandex.ru/"

	tests := map[string]struct {
		id              string
		serviceResponse string
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
		"not found": {
			id:             "d8398Sj3",
			serviceError:   service.ErrNotFound,
			wantError:      true,
			wantStatusCode: http.StatusNotFound,
		},
		"deleted": {
			id:             "d8398Sj3",
			serviceError:   service.ErrGone,
			wantError:      true,
			wantStatusCode: http.StatusGone,
		},
		"success": {
			id:              "d8398Sj3",
			serviceResponse: responseURL,
			wantStatusCode:  http.StatusTemporaryRedirect,
			wantResponse:    responseURL,
		},
	}

	for tn, tt := range tests {
		t.Run(tn, func(t *testing.T) {
			t.Parallel()

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

			service := mocks.NewURLService(t)
			service.On("GetOriginalURL", ctx, tt.id).Maybe().Return(tt.serviceResponse, tt.serviceError)

			h := NewURL(service, dummyLogger)

			h.GetOriginalURL(w, r)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.wantStatusCode, res.StatusCode)

			if !tt.wantError {
				assert.Equal(t, tt.wantResponse, res.Header.Get("location"))
			}
		})
	}
}

func TestHandler_CreateShortURL(t *testing.T) {
	dummyLogger := &logger.Logger{
		Logger: slog.New(slog.NewJSONHandler(io.Discard, nil)),
	}

	userID := uuid.New()
	url := "http://yandex.ru/"
	responseURL := "http://localhost:8080/d8398Sj3"

	tests := map[string]struct {
		ctx              context.Context
		body             io.Reader
		url              string
		readBodyResponse int
		readBodyError    error
		serviceResponse  string
		serviceError     error
		wantError        bool
		wantContentType  string
		wantStatusCode   int
		wantResponse     []byte
	}{
		"failed to get user id from context": {
			ctx:            context.Background(),
			wantError:      true,
			wantStatusCode: http.StatusInternalServerError,
		},
		"failed to read body": {
			ctx:            auth.SetUserIDToContext(context.Background(), userID),
			body:           &failReader{},
			readBodyError:  errors.New("read body error"),
			wantError:      true,
			wantStatusCode: http.StatusBadRequest,
		},
		"service error": {
			ctx:              auth.SetUserIDToContext(context.Background(), userID),
			body:             strings.NewReader(url),
			readBodyResponse: 0,
			serviceError:     errors.New("service error"),
			wantError:        true,
			wantStatusCode:   http.StatusInternalServerError,
		},
		"service error conflict": {
			ctx:              auth.SetUserIDToContext(context.Background(), userID),
			body:             strings.NewReader(url),
			readBodyResponse: 0,
			serviceError:     service.ErrConflict,
			serviceResponse:  responseURL,
			wantStatusCode:   http.StatusConflict,
			wantContentType:  "text/plain",
			wantResponse:     []byte(responseURL),
		},
		"success": {
			ctx:              auth.SetUserIDToContext(context.Background(), userID),
			body:             strings.NewReader(url),
			readBodyResponse: 0,
			serviceResponse:  responseURL,
			wantStatusCode:   http.StatusCreated,
			wantContentType:  "text/plain",
			wantResponse:     []byte(responseURL),
		},
	}
	for tn, tt := range tests {
		t.Run(tn, func(t *testing.T) {
			t.Parallel()

			r := httptest.NewRequest(http.MethodPost, "/", tt.body)
			r = r.WithContext(tt.ctx)
			w := httptest.NewRecorder()

			s := mocks.NewURLService(t)
			dto := dto.NewCreateShortURL(url, userID)
			s.On("CreateShortURL", r.Context(), dto).Maybe().Return(tt.serviceResponse, tt.serviceError)

			h := NewURL(s, dummyLogger)

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

func TestHandler_CreateShortURLJSON(t *testing.T) {
	dummyLogger := &logger.Logger{
		Logger: slog.New(slog.NewJSONHandler(io.Discard, nil)),
	}

	userID := uuid.New()
	url := "http://yandex.ru/"
	responseURL := "http://localhost:8080/d8398Sj3"

	tests := map[string]struct {
		ctx             context.Context
		body            string
		serviceResponse string
		serviceError    error
		wantError       bool
		wantStatusCode  int
		wantContentType string
		wantResponse    string
	}{
		"failed to ger user id from context": {
			ctx:            context.Background(),
			wantError:      true,
			wantStatusCode: http.StatusInternalServerError,
		},
		"failed to decode body": {
			ctx:            auth.SetUserIDToContext(context.Background(), userID),
			body:           "wrong body",
			wantError:      true,
			wantStatusCode: http.StatusBadRequest,
		},
		"service error": {
			ctx:            auth.SetUserIDToContext(context.Background(), userID),
			body:           fmt.Sprintf(`{"url": "%s"}`, url),
			serviceError:   errors.New("service error"),
			wantError:      true,
			wantStatusCode: http.StatusInternalServerError,
		},
		"service error conflict": {
			ctx:             auth.SetUserIDToContext(context.Background(), userID),
			body:            fmt.Sprintf(`{"url": "%s"}`, url),
			serviceError:    service.ErrConflict,
			serviceResponse: responseURL,
			wantStatusCode:  http.StatusConflict,
			wantContentType: "application/json",
			wantResponse:    fmt.Sprintf(`{"result": "%s"}`, responseURL),
		},
		"success": {
			ctx:             auth.SetUserIDToContext(context.Background(), userID),
			body:            fmt.Sprintf(`{"url": "%s"}`, url),
			serviceResponse: responseURL,
			wantStatusCode:  http.StatusCreated,
			wantContentType: "application/json",
			wantResponse:    fmt.Sprintf(`{"result": "%s"}`, responseURL),
		},
	}

	for tn, tt := range tests {
		t.Run(tn, func(t *testing.T) {
			t.Parallel()

			r := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(tt.body))
			r = r.WithContext(tt.ctx)
			w := httptest.NewRecorder()

			s := mocks.NewURLService(t)
			dto := dto.NewCreateShortURL(url, userID)
			s.On("CreateShortURL", r.Context(), dto).Maybe().Return(tt.serviceResponse, tt.serviceError)

			h := NewURL(s, dummyLogger)

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

func TestHandler_CreateShortURLBatch(t *testing.T) {
	dummyLogger := &logger.Logger{
		Logger: slog.New(slog.NewJSONHandler(io.Discard, nil)),
	}

	userID := uuid.New()

	tests := map[string]struct {
		ctx             context.Context
		body            string
		serviceRequest  []*request.CreateShortURLBatch
		serviceResponse []*response.CreateShortURLBatch
		serviceError    error
		wantError       bool
		wantStatusCode  int
		wantContentType string
		wantResponse    string
	}{
		"failed to get user id from context": {
			ctx:            context.Background(),
			wantError:      true,
			wantStatusCode: http.StatusInternalServerError,
		},
		"failed to decode body": {
			ctx:            auth.SetUserIDToContext(context.Background(), userID),
			body:           "wrong body",
			wantError:      true,
			wantStatusCode: http.StatusBadRequest,
		},
		"empty batch": {
			ctx:            auth.SetUserIDToContext(context.Background(), userID),
			body:           "[]",
			wantError:      true,
			wantStatusCode: http.StatusBadRequest,
		},
		"service error": {
			ctx:  auth.SetUserIDToContext(context.Background(), userID),
			body: `[{"correlation_id": "1", "original_url": "http://yandex.ru/"}, {"correlation_id": "2", "original_url": "http://google.com"}]`,
			serviceRequest: []*request.CreateShortURLBatch{
				{
					CorrelationID: "1",
					OriginalURL:   "http://yandex.ru/",
				},
				{
					CorrelationID: "2",
					OriginalURL:   "http://google.com",
				},
			},
			serviceError:   errors.New("service error"),
			wantError:      true,
			wantStatusCode: http.StatusInternalServerError,
		},
		"success": {
			ctx:  auth.SetUserIDToContext(context.Background(), userID),
			body: `[{"correlation_id": "1", "original_url": "http://yandex.ru/"}, {"correlation_id": "2", "original_url": "http://google.com"}]`,
			serviceRequest: []*request.CreateShortURLBatch{
				{
					CorrelationID: "1",
					OriginalURL:   "http://yandex.ru/",
				},
				{
					CorrelationID: "2",
					OriginalURL:   "http://google.com",
				},
			},
			serviceResponse: []*response.CreateShortURLBatch{
				{
					CorrelationID: "1",
					ShortURL:      "http://localhost:8000/yndx",
				},
				{
					CorrelationID: "2",
					ShortURL:      "http://localhost:8000/ggl",
				},
			},
			wantStatusCode:  http.StatusCreated,
			wantContentType: "application/json",
			wantResponse:    `[{"correlation_id": "1", "short_url": "http://localhost:8000/yndx"}, {"correlation_id": "2", "short_url": "http://localhost:8000/ggl"}]`,
		},
	}

	for tn, tt := range tests {
		t.Run(tn, func(t *testing.T) {
			t.Parallel()

			r := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(tt.body))
			r = r.WithContext(tt.ctx)
			w := httptest.NewRecorder()

			s := mocks.NewURLService(t)
			dto := dto.NewCreateShortURLBatch(tt.serviceRequest, userID)
			s.On("CreateShortURLBatch", r.Context(), dto).Maybe().Return(tt.serviceResponse, tt.serviceError)

			h := NewURL(s, dummyLogger)

			h.CreateShortURLBatch(w, r)

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

func TestHandler_GetUserURLs(t *testing.T) {
	dummyLogger := &logger.Logger{
		Logger: slog.New(slog.NewJSONHandler(io.Discard, nil)),
	}

	userID := uuid.New()

	tests := map[string]struct {
		ctx             context.Context
		serviceResponse []*response.GetUserURL
		serviceError    error
		wantError       bool
		wantStatusCode  int
		wantResponse    string
	}{
		"failed to get user id from context": {
			ctx:            context.Background(),
			wantError:      true,
			wantStatusCode: http.StatusInternalServerError,
		},
		"service error": {
			ctx:            auth.SetUserIDToContext(context.Background(), userID),
			serviceError:   errors.New("service error"),
			wantError:      true,
			wantStatusCode: http.StatusInternalServerError,
		},
		"service error no content": {
			ctx:            auth.SetUserIDToContext(context.Background(), userID),
			serviceError:   service.ErrNoContent,
			wantError:      true,
			wantStatusCode: http.StatusNoContent,
		},
		"success": {
			ctx: auth.SetUserIDToContext(context.Background(), userID),
			serviceResponse: []*response.GetUserURL{
				{
					ShortURL:    "http://localhost/ABOBA",
					OriginalURL: "http://yandex.ru",
				},
			},
			wantError:      false,
			wantStatusCode: http.StatusOK,
			wantResponse:   `[{"short_url": "http://localhost/ABOBA", "original_url": "http://yandex.ru"}]`,
		},
	}

	for tn, tt := range tests {
		t.Run(tn, func(t *testing.T) {
			t.Parallel()

			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r = r.WithContext(tt.ctx)

			w := httptest.NewRecorder()

			service := mocks.NewURLService(t)
			service.On("GetUserURLs", tt.ctx, userID).Maybe().Return(tt.serviceResponse, tt.serviceError)

			h := NewURL(service, dummyLogger)

			h.GetUserURLs(w, r)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.wantStatusCode, res.StatusCode)

			if !tt.wantError {
				resBody, err := io.ReadAll(res.Body)
				require.NoError(t, err)

				assert.JSONEq(t, tt.wantResponse, string(resBody))
			}
		})
	}
}

func TestHandler_DeleteURLs(t *testing.T) {
	dummyLogger := &logger.Logger{
		Logger: slog.New(slog.NewJSONHandler(io.Discard, nil)),
	}

	shortKeys := []string{"ggl", "ydx"}
	shortKeysBytes, err := json.Marshal(shortKeys)
	require.NoError(t, err)
	userID := uuid.New()

	tests := map[string]struct {
		ctx            context.Context
		body           string
		serviceError   error
		wantStatusCode int
	}{
		"failed to get user id from context": {
			ctx:            context.Background(),
			wantStatusCode: http.StatusInternalServerError,
		},
		"failed to decode body": {
			ctx:            auth.SetUserIDToContext(context.Background(), userID),
			body:           "fail",
			wantStatusCode: http.StatusBadRequest,
		},
		"service error": {
			ctx:            auth.SetUserIDToContext(context.Background(), userID),
			body:           string(shortKeysBytes),
			serviceError:   errors.New("service error"),
			wantStatusCode: http.StatusInternalServerError,
		},
		"success": {
			ctx:            auth.SetUserIDToContext(context.Background(), userID),
			body:           string(shortKeysBytes),
			wantStatusCode: http.StatusAccepted,
		},
	}

	for tn, tt := range tests {
		t.Run(tn, func(t *testing.T) {
			t.Parallel()

			r := httptest.NewRequest(http.MethodDelete, "/", strings.NewReader(tt.body))
			r = r.WithContext(tt.ctx)

			w := httptest.NewRecorder()

			serviceMock := mocks.NewURLService(t)
			dto := dto.NewDeleteURLs(shortKeys, userID)
			serviceMock.On("DeleteURLs", tt.ctx, dto).Maybe().
				Return(tt.serviceError)

			h := NewURL(serviceMock, dummyLogger)

			h.DeleteURLs(w, r)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.wantStatusCode, res.StatusCode)

			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Empty(t, resBody)
		})
	}
}
