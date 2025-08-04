package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/dtroode/urlshorter/internal/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// RequestLog represents request logging middleware.
// It provides structured logging for HTTP and gRPC requests, including timing and response metrics.
type RequestLog struct {
	l *logger.Logger
}

// NewRequestLog creates new RequestLog instance.
// It initializes request logging middleware with the provided logger.
//
// Parameters:
//   - l: The logger for recording request information
//
// Returns a pointer to the newly created RequestLog middleware instance.
func NewRequestLog(l *logger.Logger) *RequestLog {
	return &RequestLog{
		l: l,
	}
}

// HandleHTTP processes HTTP requests and logs them.
// It wraps HTTP requests with timing and response metrics logging.
// Logs include URI, method, status code, duration, and response size.
//
// Parameters:
//   - h: The next HTTP handler in the middleware chain
//
// Returns an HTTP handler that logs request information before and after processing.
func (m *RequestLog) HandleHTTP(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		h.ServeHTTP(&lw, r)

		duration := time.Since(start)

		m.l.Info(
			"request",
			"uri", r.RequestURI,
			"method", r.Method,
			"status", lw.responseData.status,
			"duration", duration,
			"size", lw.responseData.size,
		)
	})
}

// HandleGRPC processes GRPC requests and logs them.
// It wraps gRPC requests with timing and response metrics logging.
// Logs include method name, status code, duration, and response size.
//
// Parameters:
//   - ctx: The gRPC context
//   - req: The gRPC request
//   - info: The gRPC server info containing method information
//   - handler: The next gRPC handler in the interceptor chain
//
// Returns the gRPC response and any error that occurred during processing.
func (m *RequestLog) HandleGRPC(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()

	response, err := handler(ctx, req)

	responseSize := 0
	if resp, ok := response.(proto.Message); ok {
		responseSize = proto.Size(resp)
	}

	duration := time.Since(start)

	m.l.Info(
		"request",
		"uri", info.FullMethod,
		"method", info.FullMethod,
		"status", status.FromContextError(err).Code(),
		"duration", duration,
		"size", responseSize,
	)

	return response, err
}

type responseData struct {
	status int
	size   int
}

type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

// WriteHeader captures response status code for logging.
func (w *loggingResponseWriter) WriteHeader(statusCode int) {
	w.responseData.status = statusCode

	w.ResponseWriter.WriteHeader(statusCode)
}

// Write captures response size for logging.
func (w *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := w.ResponseWriter.Write(b)
	w.responseData.size = size
	return size, err
}
