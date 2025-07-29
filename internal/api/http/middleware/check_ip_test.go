package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckIP(t *testing.T) {
	tests := []struct {
		name             string
		trustedSubnet    string
		remoteAddr       string
		header           string
		headerValue      string
		expectedStatus   int
		expectNewErr     bool
		expectHandlerRun bool
	}{
		{
			name:             "allowed ip from X-Real-IP",
			trustedSubnet:    "192.168.1.0/24",
			remoteAddr:       "10.0.0.1:12345",
			header:           "X-Real-IP",
			headerValue:      "192.168.1.10",
			expectedStatus:   http.StatusOK,
			expectHandlerRun: true,
		},
		{
			name:             "allowed ip from X-Forwarded-For",
			trustedSubnet:    "192.168.1.0/24",
			remoteAddr:       "10.0.0.1:12345",
			header:           "X-Forwarded-For",
			headerValue:      "192.168.1.10, 10.0.0.5",
			expectedStatus:   http.StatusOK,
			expectHandlerRun: true,
		},
		{
			name:           "forbidden ip",
			trustedSubnet:  "192.168.1.0/24",
			remoteAddr:     "10.0.0.1:12345",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "invalid trusted subnet",
			trustedSubnet:  "invalid",
			remoteAddr:     "192.168.1.10:12345",
			expectedStatus: http.StatusInternalServerError,
			expectNewErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkIP, err := NewCheckIP(tt.trustedSubnet)
			if tt.expectNewErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			handlerCalled := false
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.header != "" {
				req.Header.Set(tt.header, tt.headerValue)
			}

			w := httptest.NewRecorder()
			middleware := checkIP.Handle(handler)
			middleware.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectHandlerRun, handlerCalled)
		})
	}
}
