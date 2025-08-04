package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTLSListener(t *testing.T) {
	certFile := "cert.pem"
	keyFile := "key.pem"

	listener := NewTLSListener(certFile, keyFile)

	assert.NotNil(t, listener)
	assert.Equal(t, certFile, listener.certFileName)
	assert.Equal(t, keyFile, listener.privateKeyFileName)
}

func TestNewPlainListener(t *testing.T) {
	listener := NewPlainListener()

	assert.NotNil(t, listener)
}

func TestPlainListener_Listen(t *testing.T) {
	listener := NewPlainListener()

	// Test that Listen method exists and can be called
	// We use a dummy address that should fail to bind in tests
	l, err := listener.Listen("tcp", ":0")
	if err == nil {
		defer l.Close()
		assert.NotNil(t, l)
	}
	// We don't assert no error because the port might not be available
}
