package server

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHTTPServer(t *testing.T) {
	httpServer := &http.Server{}
	addr := ":8080"

	server := NewHTTPServer(httpServer, addr)

	assert.NotNil(t, server)
	assert.Equal(t, httpServer, server.server)
	assert.Equal(t, addr, server.addr)
}

func TestHTTPServer_Address(t *testing.T) {
	httpServer := &http.Server{}
	addr := ":8080"

	server := NewHTTPServer(httpServer, addr)

	assert.Equal(t, addr, server.Address())
}
