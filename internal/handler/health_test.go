package handler

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dtroode/urlshorter/internal/handler/mocks"
	"github.com/dtroode/urlshorter/internal/logger"
)

func TestHealth_Ping(t *testing.T) {
	dummyLogger := &logger.Logger{
		Logger: slog.New(slog.NewJSONHandler(io.Discard, nil)),
	}

	tests := map[string]struct {
		serviceError       error
		expectedStatusCode int
	}{
		"error from service": {
			serviceError:       errors.New("service error"),
			expectedStatusCode: http.StatusInternalServerError,
		},
		"nil from service": {
			expectedStatusCode: http.StatusOK,
		},
	}

	for tn, tt := range tests {
		tt := tt

		t.Run(tn, func(t *testing.T) {
			t.Parallel()

			r := httptest.NewRequest(http.MethodGet, "/ping", nil)
			ctx := r.Context()
			w := httptest.NewRecorder()

			serviceMock := mocks.NewHealthService(t)
			serviceMock.On("Ping", ctx).Once().Return(tt.serviceError)

			h := NewHealth(serviceMock, dummyLogger)

			h.Ping(w, r)

			require.Equal(t, tt.expectedStatusCode, w.Code)
		})
	}
}
