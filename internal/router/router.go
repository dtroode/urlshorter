package router

import (
	"github.com/dtroode/urlshorter/internal/handler"
	"github.com/dtroode/urlshorter/internal/logger"
	"github.com/dtroode/urlshorter/internal/middleware"
	"github.com/dtroode/urlshorter/internal/service"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

type Router struct {
	chi.Router
}

func NewRouter() *Router {
	return &Router{
		Router: chi.NewRouter(),
	}
}

func (r *Router) RegisterRoutes(s *service.URL, logger *logger.Logger) {
	loggerMiddleware := middleware.NewRequestLog(logger).Handle
	degzipper := middleware.Decompress
	compressor := chiMiddleware.Compress(5)

	h := handler.NewHandler(s)

	r.Route("/", func(r chi.Router) {
		r.Use(loggerMiddleware)
		r.Use(degzipper)

		r.With(compressor).Post("/", h.CreateShortURL)
		r.Get("/{id}", h.GetShortURL)
		r.With(compressor).Post("/", h.CreateShortURL)
		r.Get("/{id}", h.GetShortURL)
	})

	r.Route("/api", func(r chi.Router) {
		r.Use(loggerMiddleware)
		r.Use(degzipper)
		r.Use(compressor)

		r.Post("/shorten", h.CreateShortURLJSON)
	})
}
