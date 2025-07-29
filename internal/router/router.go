package router

import (
	"fmt"

	"github.com/go-chi/chi/v5"

	"github.com/dtroode/urlshorter/internal/api/http/handler"
	"github.com/dtroode/urlshorter/internal/api/http/middleware"
	"github.com/dtroode/urlshorter/internal/logger"
	"github.com/dtroode/urlshorter/internal/service"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

// Router represents HTTP router with route registration capabilities.
type Router struct {
	chi.Router
}

// NewRouter creates new Router instance.
func NewRouter() *Router {
	return &Router{
		Router: chi.NewRouter(),
	}
}

// RegisterProfiler registers profiler routes for debugging.
func (r *Router) RegisterProfiler() {
	r.Mount("/debug", chiMiddleware.Profiler())
}

// RegisterAPIRoutes registers API routes with middleware.
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
		r.Get("/{id}", h.GetOriginalURL)
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
			r.Delete("/urls", h.DeleteURLs)
		})
	})
}

// RegisterHealthRoutes registers health check routes.
func (r *Router) RegisterHealthRoutes(s *service.Health, l *logger.Logger) {
	loggerMiddleware := middleware.NewRequestLog(l).Handle
	h := handler.NewHealth(s, l)

	r.With(loggerMiddleware).Get("/ping", h.Ping)
}

// RegisterInternalRoutes registers internal API routes.
func (r *Router) RegisterInternalRoutes(s *service.URL, l *logger.Logger, cidr string) error {
	loggerMiddleware := middleware.NewRequestLog(l).Handle
	checkIPMiddleware, err := middleware.NewCheckIP(cidr)
	if err != nil {
		return fmt.Errorf("failed to register check ip middleware")
	}
	h := handler.NewInternal(s, l)

	r.With(loggerMiddleware).With(checkIPMiddleware.Handle).Get("/api/internal/stats", h.GetStats)

	return nil
}
