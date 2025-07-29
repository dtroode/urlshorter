package middleware

import (
	"net/http"
	"time"

	"github.com/dtroode/urlshorter/internal/logger"
)

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

// RequestLog represents request logging middleware.
type RequestLog struct {
	l *logger.Logger
}

// NewRequestLog creates new RequestLog instance.
func NewRequestLog(l *logger.Logger) *RequestLog {
	return &RequestLog{
		l: l,
	}
}

// Handle processes HTTP requests and logs them.
func (m *RequestLog) Handle(h http.Handler) http.Handler {
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
