package server

import (
	"context"

	"github.com/dtroode/urlshorter/internal/model"
	"google.golang.org/grpc"
)

// GRPCServer represents a gRPC server that can be started and stopped gracefully.
// It wraps the standard grpc.Server with additional functionality for graceful shutdown.
type GRPCServer struct {
	server *grpc.Server
	addr   string
}

// NewGRPCServer creates a new GRPCServer instance.
// It wraps the provided grpc.Server with the specified address for graceful management.
//
// Parameters:
//   - server: The grpc.Server to wrap
//   - addr: The address the server will listen on
//
// Returns a pointer to the newly created GRPCServer instance.
func NewGRPCServer(
	server *grpc.Server,
	addr string,
) *GRPCServer {
	return &GRPCServer{
		server: server,
		addr:   addr,
	}
}

// Start starts the gRPC server using the provided security layer.
// It creates a listener using the security layer and begins serving gRPC requests.
//
// Parameters:
//   - securityLayer: The security layer to use for creating the listener
//
// Returns an error if the server fails to start.
func (s *GRPCServer) Start(securityLayer model.SecurityLayer) error {
	listener, err := securityLayer.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	return s.server.Serve(listener)
}

// Stop gracefully shuts down the gRPC server.
// It waits for ongoing requests to complete before shutting down.
//
// Parameters:
//   - ctx: The context with timeout for the shutdown operation
//
// Returns an error if the shutdown fails.
func (s *GRPCServer) Stop(ctx context.Context) error {
	s.server.GracefulStop()
	return nil
}

// Address returns the address the server is configured to listen on.
//
// Returns the server address as a string.
func (s *GRPCServer) Address() string {
	return s.addr
}
