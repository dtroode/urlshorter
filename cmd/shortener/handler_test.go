package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dtroode/urlshorter/cmd/shortener/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type failReader struct {
}

func (m *failReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("failed to read")
}

func TestCreateShortURLHandler_ServeHTTP(t *testing.T) {
	ctx := context.Background()

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
			service.On("CreateShortURL", ctx, url).Maybe().Return(tt.serviceResponse, tt.serviceError)

			h := NewCreateShorURLHandler(service)

			h.ServeHTTP(w, r)

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

func TestGetShortURLHandler_ServeHTTP(t *testing.T) {
	ctx := context.Background()

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
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.SetPathValue("id", tt.id)
			w := httptest.NewRecorder()

			service := mocks.NewService(t)
			service.On("GetOriginalURL", ctx, tt.id).Maybe().Return(tt.serviceResponse, tt.serviceError)

			h := NewGetShortURLHandler(service)

			h.ServeHTTP(w, r)

			res := w.Result()

			assert.Equal(t, tt.wantStatusCode, res.StatusCode)

			if !tt.wantError {
				assert.Equal(t, tt.wantResponse, res.Header.Get("location"))
			}
		})
	}
}
