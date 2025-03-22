package router

import (
	"github.com/dtroode/urlshorter/internal/handler"
	internalLogger "github.com/dtroode/urlshorter/internal/logger"
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

func (r *Router) RegisterRoutes(s *service.URL, logger *internalLogger.Log) {
	myLogFormatter := internalLogger.NewLogFormatter(*logger)
	requestLogger := chiMiddleware.RequestLogger(myLogFormatter)

	degzipper := middleware.DeGzip
	compressor := chiMiddleware.Compress(5)

	h := handler.NewHandler(s)

	r.Route("/", func(r chi.Router) {
		r.Use(requestLogger)
		r.Use(degzipper)

		r.With(compressor).Post("/", h.CreateShortURL)
		r.Get("/{id}", h.GetShortURL)
	})
}

func (r *Router) RegisterAPIRoutes(s *service.URL, logger *internalLogger.Log) {
	myLogFormatter := internalLogger.NewLogFormatter(*logger)
	requestLogger := chiMiddleware.RequestLogger(myLogFormatter)

	degzipper := middleware.DeGzip
	compressor := chiMiddleware.Compress(5)

	h := handler.NewHandlerJSON(s)

	r.Route("/api", func(r chi.Router) {
		r.Use(requestLogger)
		r.Use(degzipper)
		r.Use(compressor)

		r.Post("/shorten", h.CreateShortURL)
	})
}
