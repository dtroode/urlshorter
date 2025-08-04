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

func TestNew(t *testing.T) {
	mockStorage := mocks.NewURLStorage(t)
	urlService := service.NewURL("http://localhost:8080", 8, 3, 0, mockStorage)
	healthService := &service.Health{}
	token := auth.NewJWT("test-secret")
	logger := &logger.Logger{Logger: slog.New(slog.NewJSONHandler(io.Discard, nil))}

	router := New(urlService, healthService, token, logger, "192.168.1.0/24")
	assert.NotNil(t, router)
	assert.NotNil(t, router.Router)
}

func TestRouter_Register(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockStorage := mocks.NewURLStorage(t)
		urlService := service.NewURL("http://localhost:8080", 8, 3, 0, mockStorage)
		healthService := &service.Health{}
		token := auth.NewJWT("test-secret")
		logger := &logger.Logger{Logger: slog.New(slog.NewJSONHandler(io.Discard, nil))}

		router := New(urlService, healthService, token, logger, "192.168.1.0/24")
		err := router.Register()
		assert.NoError(t, err)
	})

	t.Run("invalid cidr", func(t *testing.T) {
		mockStorage := mocks.NewURLStorage(t)
		urlService := service.NewURL("http://localhost:8080", 8, 3, 0, mockStorage)
		healthService := &service.Health{}
		token := auth.NewJWT("test-secret")
		logger := &logger.Logger{Logger: slog.New(slog.NewJSONHandler(io.Discard, nil))}

		router := New(urlService, healthService, token, logger, "invalid-cidr")
		err := router.Register()
		assert.Error(t, err)
	})
}
