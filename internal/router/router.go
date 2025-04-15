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

func (r *Router) RegisterAPIRoutes(s *service.URL, token middleware.Token, l *logger.Logger) {
	loggerMiddleware := middleware.NewRequestLog(l).Handle
	authenticate := middleware.NewAuthenticate(token, l).Handle
	degzipper := middleware.Decompress
	compressor := chiMiddleware.Compress(5)

	h := handler.NewURL(s, l)

	r.Route("/", func(r chi.Router) {
		r.Use(loggerMiddleware)
		r.Use(authenticate)
		r.Use(degzipper)

		r.With(compressor).Post("/", h.CreateShortURL)
		r.Get("/{id}", h.GetShortURL)
		r.With(compressor).Post("/", h.CreateShortURL)
		r.Get("/{id}", h.GetShortURL)
	})

	r.Route("/api", func(r chi.Router) {
		r.Use(loggerMiddleware)
		r.Use(authenticate)
		r.Use(degzipper)
		r.Use(compressor)

		r.Route("/shorten", func(r chi.Router) {
			r.Post("/", h.CreateShortURLJSON)
			r.Post("/batch", h.CreateShortURLBatch)
		})

		r.Route("/user", func(r chi.Router) {
			r.Get("/urls", h.GetUserURLs)
		})

	})
}

func (r *Router) RegisterHealthRoutes(s *service.Health, l *logger.Logger) {
	loggerMiddleware := middleware.NewRequestLog(l).Handle
	h := handler.NewHealth(s, l)

	r.With(loggerMiddleware).Get("/ping", h.Ping)
}
