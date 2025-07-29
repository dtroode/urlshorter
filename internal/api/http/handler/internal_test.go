package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/dtroode/urlshorter/internal/logger"
	"github.com/dtroode/urlshorter/internal/model"
)

type MockStatsService struct {
	mock.Mock
}

func (m *MockStatsService) GetStats(ctx context.Context) (model.GetStats, error) {
	args := m.Called(ctx)
	return args.Get(0).(model.GetStats), args.Error(1)
}

func TestInternal_GetStats(t *testing.T) {
	tests := []struct {
		name           string
		serviceStats   model.GetStats
		serviceErr     error
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success",
			serviceStats: model.GetStats{
				URLs:  10,
				Users: 5,
			},
			serviceErr:     nil,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"urls":10,"users":5}` + "\n",
		},
		{
			name:           "service error",
			serviceStats:   model.GetStats{},
			serviceErr:     errors.New("service error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockStatsService)
			mockService.On("GetStats", mock.Anything).Return(tt.serviceStats, tt.serviceErr)

			log := logger.NewLog("info")
			handler := NewInternal(mockService, log)

			r := chi.NewRouter()
			r.Get("/api/internal/stats", handler.GetStats)

			req := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			}
			mockService.AssertExpectations(t)
		})
	}
}
