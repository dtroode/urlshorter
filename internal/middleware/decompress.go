package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

type gzipReader struct {
	g *gzip.Reader
	r io.ReadCloser
}

// Close closes both the gzip reader and the underlying reader.
func (gr *gzipReader) Close() error {
	if err := gr.r.Close(); err != nil {
		return err
	}
	return gr.g.Close()
}

// Read reads data from the gzip reader.
func (gr *gzipReader) Read(p []byte) (n int, err error) {
	return gr.g.Read(p)
}

var readerPool = &sync.Pool{
	New: func() any {
		return (*gzip.Reader)(nil)
	},
}

// Decompress middleware decompresses gzipped HTTP requests.
func Decompress(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ce := r.Header.Get("Content-Encoding")
		if strings.Contains(ce, "gzip") {
			gz := readerPool.Get().(*gzip.Reader)

			if gz == nil {
				var err error
				gz, err = gzip.NewReader(r.Body)
				if err != nil {
					http.Error(w, "failed to create gzip reader", http.StatusInternalServerError)
					return
				}
			} else {
				gz.Reset(r.Body)
			}
			defer readerPool.Put(gz)

			gr := &gzipReader{
				g: gz,
				r: r.Body,
			}

			r.Body = gr
		}

		h.ServeHTTP(w, r)
	})
}
