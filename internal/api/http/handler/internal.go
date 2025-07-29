package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/dtroode/urlshorter/internal/api/http/response"
	"github.com/dtroode/urlshorter/internal/logger"
	"github.com/dtroode/urlshorter/internal/model"
)

// StatsService defines the interface for the statistics service.
type StatsService interface {
	// GetStats retrieves statistics about the service.
	GetStats(ctx context.Context) (model.GetStats, error)
}

// Internal holds the dependencies for the internal handlers.
type Internal struct {
	service StatsService
	logger  *logger.Logger
}

// NewInternal creates a new Internal handler.
func NewInternal(s StatsService, l *logger.Logger) *Internal {
	return &Internal{
		service: s,
		logger:  l,
	}
}

// GetStats handles the request to get service statistics.
func (h *Internal) GetStats(w http.ResponseWriter, r *http.Request) {
	var (
		ctx   = r.Context()
		stats model.GetStats
		resp  response.GetStats
		err   error
	)

	stats, err = h.service.GetStats(ctx)
	if err != nil {
		h.logger.Error("failed to get stats", "error", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	resp = response.GetStats{
		URLs:  stats.URLs,
		Users: stats.Users,
	}

	w.Header().Set("content-type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("failed to encode response", "error", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}
}
