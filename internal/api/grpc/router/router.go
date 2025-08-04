package router

import (
	"github.com/dtroode/urlshorter/internal/api/grpc/handler"
	pb "github.com/dtroode/urlshorter/internal/api/grpc/proto"
	"github.com/dtroode/urlshorter/internal/api/middleware"
	"github.com/dtroode/urlshorter/internal/logger"
	"github.com/dtroode/urlshorter/internal/service"
	"google.golang.org/grpc"
)

// Router represents a gRPC router for URL shortening operations.
// It manages gRPC service registration and middleware configuration.
type Router struct {
	urlService   *service.URL
	tokenManager middleware.Token
	logger       *logger.Logger
}

// New creates new gRPC Router instance.
// It initializes a gRPC router with URL service and authentication middleware.
//
// Parameters:
//   - urlService: The URL shortening service
//   - tokenManager: The token management service for authentication
//   - logger: The logger for request logging
//
// Returns a pointer to the newly created Router instance.
func New(
	urlService *service.URL,
	tokenManager middleware.Token,
	logger *logger.Logger,
) *Router {
	return &Router{
		urlService:   urlService,
		tokenManager: tokenManager,
		logger:       logger,
	}
}

// Register registers all gRPC services and middleware.
// It sets up the gRPC server with request logging and authentication interceptors.
//
// Returns the configured gRPC server instance.
func (r *Router) Register() *grpc.Server {
	requestLog := middleware.NewRequestLog(r.logger).HandleGRPC
	authenticate := middleware.NewAuthenticate(r.tokenManager, r.logger).HandleGRPC

	s := grpc.NewServer(grpc.ChainUnaryInterceptor(requestLog, authenticate))
	r.registerAPIRoutes(s)

	return s
}

func (r *Router) registerAPIRoutes(server *grpc.Server) {
	h := handler.NewURL(r.urlService, r.logger)

	server.RegisterService(&pb.URLService_ServiceDesc, h)
}
