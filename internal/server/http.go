package server

import (
	"context"
	"net/http"

	"github.com/dtroode/urlshorter/internal/model"
)

// HTTPServer represents an HTTP server that can be started and stopped gracefully.
// It wraps the standard http.Server with additional functionality for graceful shutdown.
type HTTPServer struct {
	server *http.Server
	addr   string
}

// NewHTTPServer creates a new HTTPServer instance.
// It wraps the provided http.Server with the specified address for graceful management.
//
// Parameters:
//   - server: The http.Server to wrap
//   - addr: The address the server will listen on
//
// Returns a pointer to the newly created HTTPServer instance.
func NewHTTPServer(
	server *http.Server,
	addr string,
) *HTTPServer {
	return &HTTPServer{
		server: server,
		addr:   addr,
	}
}

// Start starts the HTTP server using the provided security layer.
// It creates a listener using the security layer and begins serving HTTP requests.
//
// Parameters:
//   - securityLayer: The security layer to use for creating the listener
//
// Returns an error if the server fails to start.
func (s *HTTPServer) Start(securityLayer model.SecurityLayer) error {
	listener, err := securityLayer.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	return s.server.Serve(listener)
}

// Stop gracefully shuts down the HTTP server.
// It waits for ongoing requests to complete before shutting down.
//
// Parameters:
//   - ctx: The context with timeout for the shutdown operation
//
// Returns an error if the shutdown fails.
func (s *HTTPServer) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// Address returns the address the server is configured to listen on.
//
// Returns the server address as a string.
func (s *HTTPServer) Address() string {
	return s.addr
}
