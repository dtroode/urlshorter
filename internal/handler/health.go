package handler

import (
	"context"
	"net/http"

	"github.com/dtroode/urlshorter/internal/logger"
)

type HealthService interface {
	Ping(ctx context.Context) error
}

type Health struct {
	service HealthService
	logger  *logger.Logger
}

func NewHealth(s HealthService, l *logger.Logger) *Health {
	return &Health{
		service: s,
		logger:  l,
	}
}

func (h *Health) Ping(w http.ResponseWriter, r *http.Request) {
	err := h.service.Ping(r.Context())
	if err != nil {
		h.logger.Error("ping failed", "error", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusOK)
}
