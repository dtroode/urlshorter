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
// @Summary Health check
// @Description Check if the service is running and database is accessible
// @Tags Health
// @Accept json
// @Produce json
// @Success 200 {string} string "Service is healthy"
// @Failure 500 {string} string "Service is unhealthy"
// @Router /ping [get]
func (h *Health) Ping(w http.ResponseWriter, r *http.Request) {
	err := h.service.Ping(r.Context())
	if err != nil {
		h.logger.Error("ping failed", "error", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusOK)
}
