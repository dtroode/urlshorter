package router

import (
	"github.com/dtroode/urlshorter/internal/handler"
	"github.com/dtroode/urlshorter/internal/middleware"
	"github.com/dtroode/urlshorter/internal/service"
	"github.com/go-chi/chi/v5"
)

type Router struct {
	chi.Router
}

func NewRouter() *Router {
	return &Router{
		Router: chi.NewRouter(),
	}
}

func (r *Router) RegisterRoutes(s *service.URL) {
	r.Route("/", func(r chi.Router) {
		h := handler.NewHandler(s)

		r.Post("/", middleware.RequestLog(h.CreateShortURL))
		r.Get("/{id}", middleware.RequestLog(h.GetShortURL))
	})
}

func (r *Router) RegisterAPIRoutes(s *service.URL) {
	r.Route("/api", func(r chi.Router) {
		h := handler.NewHandlerJSON(s)
		r.Post("/shorten", middleware.RequestLog(h.CreateShortURL))
	})
}
