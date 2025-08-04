package router

import (
	"fmt"

	"github.com/go-chi/chi/v5"

	"github.com/dtroode/urlshorter/internal/api/http/handler"
	"github.com/dtroode/urlshorter/internal/api/middleware"
	"github.com/dtroode/urlshorter/internal/logger"
	"github.com/dtroode/urlshorter/internal/service"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

// Router represents HTTP router with route registration capabilities.
type Router struct {
	chi.Router
	urlService    *service.URL
	healthService *service.Health
	tokenManager  middleware.Token
	logger        *logger.Logger
	trustedSubnet string
}

// New creates new Router instance.
// It initializes an HTTP router with all necessary services and middleware.
//
// Parameters:
//   - urlService: The URL shortening service
//   - healthService: The health check service
//   - tokenManager: The token management service for authentication
//   - logger: The logger for request logging
//   - trustedSubnet: The trusted subnet for internal API access
//
// Returns a pointer to the newly created Router instance.
func New(
	urlService *service.URL,
	healthService *service.Health,
	tokenManager middleware.Token,
	logger *logger.Logger,
	trustedSubnet string,
) *Router {
	return &Router{
		Router:        chi.NewRouter(),
		urlService:    urlService,
		healthService: healthService,
		tokenManager:  tokenManager,
		logger:        logger,
		trustedSubnet: trustedSubnet,
	}
}

// Register registers all routes and middleware for the HTTP router.
// It sets up API routes, health check routes, internal routes, and debugging routes.
//
// Returns an error if route registration fails.
func (r *Router) Register() error {
	var err error

	r.registerProfiler()
	r.registerHealthRoutes()
	r.registerAPIRoutes()
	err = r.registerInternalRoutes()

	return err
}

// registerProfiler registers profiler routes for debugging.
func (r *Router) registerProfiler() {
	r.Mount("/debug", chiMiddleware.Profiler())
}

// registerAPIRoutes registers API routes with middleware.
func (r *Router) registerAPIRoutes() {
	requestLog := middleware.NewRequestLog(r.logger).HandleHTTP
	authenticate := middleware.NewAuthenticate(r.tokenManager, r.logger).HandleHTTP
	degzipper := middleware.Decompress
	compressor := chiMiddleware.Compress(5)

	h := handler.NewURL(r.urlService, r.logger)

	r.Route("/", func(r chi.Router) {
		r.Use(requestLog)
		r.Use(authenticate)
		r.Use(degzipper)

		r.With(compressor).Post("/", h.CreateShortURL)
		r.Get("/{id}", h.GetOriginalURL)
	})

	r.Route("/api", func(r chi.Router) {
		r.Use(requestLog)
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

// registerHealthRoutes registers health check routes.
func (r *Router) registerHealthRoutes() {
	requestLog := middleware.NewRequestLog(r.logger).HandleHTTP
	h := handler.NewHealth(r.healthService, r.logger)

	r.With(requestLog).Get("/ping", h.Ping)
}

// registerInternalRoutes registers internal API routes.
func (r *Router) registerInternalRoutes() error {
	requestLog := middleware.NewRequestLog(r.logger).HandleHTTP
	checkIPMiddleware, err := middleware.NewCheckIP(r.trustedSubnet)
	if err != nil {
		return fmt.Errorf("failed to register check ip middleware")
	}
	h := handler.NewInternal(r.urlService, r.logger)

	r.With(requestLog).With(checkIPMiddleware.Handle).Get("/api/internal/stats", h.GetStats)

	return nil
}
