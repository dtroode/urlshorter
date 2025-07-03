package router

import (
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dtroode/urlshorter/internal/auth"
	"github.com/dtroode/urlshorter/internal/logger"
	"github.com/dtroode/urlshorter/internal/service"
	"github.com/dtroode/urlshorter/internal/service/mocks"
)

func TestNewRouter(t *testing.T) {
	router := NewRouter()
	assert.NotNil(t, router)
	assert.NotNil(t, router.Router)
}

func TestRouter_RegisterProfiler(t *testing.T) {
	router := NewRouter()
	assert.NotPanics(t, func() {
		router.RegisterProfiler()
	})
}

func TestRouter_RegisterAPIRoutes(t *testing.T) {
	router := NewRouter()
	mockStorage := mocks.NewURLStorage(t)
	urlService := service.NewURL("http://localhost:8080", 8, 3, 0, mockStorage)
	token := auth.NewJWT("test-secret")
	logger := &logger.Logger{Logger: slog.New(slog.NewJSONHandler(io.Discard, nil))}
	assert.NotPanics(t, func() {
		router.RegisterAPIRoutes(urlService, token, logger)
	})
}

func TestRouter_RegisterHealthRoutes(t *testing.T) {
	router := NewRouter()
	healthService := &service.Health{}
	logger := &logger.Logger{Logger: slog.New(slog.NewJSONHandler(io.Discard, nil))}
	assert.NotPanics(t, func() {
		router.RegisterHealthRoutes(healthService, logger)
	})
}

func TestRouter_CompleteSetup(t *testing.T) {
	router := NewRouter()
	mockStorage := mocks.NewURLStorage(t)
	urlService := service.NewURL("http://localhost:8080", 8, 3, 0, mockStorage)
	token := auth.NewJWT("test-secret")
	logger := &logger.Logger{Logger: slog.New(slog.NewJSONHandler(io.Discard, nil))}
	healthService := &service.Health{}
	assert.NotPanics(t, func() {
		router.RegisterProfiler()
		router.RegisterAPIRoutes(urlService, token, logger)
		router.RegisterHealthRoutes(healthService, logger)
	})
}
