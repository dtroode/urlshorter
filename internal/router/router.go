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

func (r *Router) RegisterAPIRoutes(s *service.URL, l *logger.Logger) {
	loggerMiddleware := middleware.NewRequestLog(l).Handle
	degzipper := middleware.Decompress
	compressor := chiMiddleware.Compress(5)

	h := handler.NewURL(s, l)

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

func (r *Router) RegisterHealthRoutes(s *service.Health, l *logger.Logger) {
	loggerMiddleware := middleware.NewRequestLog(l).Handle
	h := handler.NewHealth(s, l)

	r.With(loggerMiddleware).Get("/ping", h.Ping)
}
