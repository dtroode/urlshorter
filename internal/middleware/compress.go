package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"
	"sync"
)

var compressibleContentTypes = map[string]bool{
	"application/javascript": true,
	"application/json":       true,
	"text/css":               true,
	"text/html":              true,
	"text/plain":             true,
	"text/xml":               true,
}

const minimumCompressSize = 1400

var writerPool = &sync.Pool{
	New: func() any {
		return gzip.NewWriter(nil)
	},
}

// Compress middleware compresses HTTP responses using gzip.
func Compress(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ae := r.Header.Get("Accept-Encoding")
		if !strings.Contains(ae, "gzip") {
			h.ServeHTTP(w, r)
			return
		}

		responseWriter := &responseCaptureWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		h.ServeHTTP(responseWriter, r)

		if !shouldCompress(responseWriter) {
			w.WriteHeader(responseWriter.statusCode)
			w.Write([]byte(responseWriter.body.String()))
			return
		}

		gz := writerPool.Get().(*gzip.Writer)
		defer writerPool.Put(gz)

		gz.Reset(w)

		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")
		w.WriteHeader(responseWriter.statusCode)

		gz.Write([]byte(responseWriter.body.String()))
		gz.Close()
	})
}

// shouldCompress determines if response should be compressed.
func shouldCompress(rw *responseCaptureWriter) bool {
	contentType := rw.Header().Get("Content-Type")
	if contentType != "" {
		mainType := strings.Split(contentType, ";")[0]
		if !compressibleContentTypes[mainType] {
			return false
		}
	}

	if rw.body.Len() < minimumCompressSize {
		return false
	}

	return true
}

type responseCaptureWriter struct {
	http.ResponseWriter
	body       strings.Builder
	statusCode int
}

// Write captures response body for compression decision.
func (rc *responseCaptureWriter) Write(data []byte) (int, error) {
	rc.body.Write(data)
	return len(data), nil
}

// WriteHeader captures response status code.
func (rc *responseCaptureWriter) WriteHeader(statusCode int) {
	rc.statusCode = statusCode
}
