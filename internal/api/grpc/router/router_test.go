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
	token := auth.NewJWT("test-secret")
	logger := &logger.Logger{Logger: slog.New(slog.NewJSONHandler(io.Discard, nil))}

	router := New(urlService, token, logger)
	assert.NotNil(t, router)
}

func TestRouter_Register(t *testing.T) {
	mockStorage := mocks.NewURLStorage(t)
	urlService := service.NewURL("http://localhost:8080", 8, 3, 0, mockStorage)
	token := auth.NewJWT("test-secret")
	logger := &logger.Logger{Logger: slog.New(slog.NewJSONHandler(io.Discard, nil))}

	router := New(urlService, token, logger)
	grpcServer := router.Register()

	assert.NotNil(t, grpcServer)
}
