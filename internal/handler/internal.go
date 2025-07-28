package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/dtroode/urlshorter/internal/logger"
	"github.com/dtroode/urlshorter/internal/response"
)

// StatsService defines the interface for the statistics service.
type StatsService interface {
	// GetStats retrieves statistics about the service.
	GetStats(ctx context.Context) (response.GetStats, error)
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
		stats response.GetStats
		err   error
	)

	stats, err = h.service.GetStats(ctx)
	if err != nil {
		h.logger.Error("failed to get stats", "error", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.Header().Set("content-type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		h.logger.Error("failed to encode response", "error", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}
}
