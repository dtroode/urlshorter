package handler

import (
	"context"
	"net/http"

	"github.com/dtroode/urlshorter/internal/logger"
)

// HealthService defines interface for health check service.
type HealthService interface {
	Ping(ctx context.Context) error
}

// Health represents handler for service health check.
type Health struct {
	service HealthService
	logger  *logger.Logger
}

// NewHealth creates new Health instance.
func NewHealth(s HealthService, l *logger.Logger) *Health {
	return &Health{
		service: s,
		logger:  l,
	}
}

// Ping handles service health check request.
func (h *Health) Ping(w http.ResponseWriter, r *http.Request) {
	err := h.service.Ping(r.Context())
	if err != nil {
		h.logger.Error("ping failed", "error", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusOK)
}
