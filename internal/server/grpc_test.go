package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestNewGRPCServer(t *testing.T) {
	grpcServer := grpc.NewServer()
	addr := ":8081"

	server := NewGRPCServer(grpcServer, addr)

	assert.NotNil(t, server)
	assert.Equal(t, grpcServer, server.server)
	assert.Equal(t, addr, server.addr)
}

func TestGRPCServer_Address(t *testing.T) {
	grpcServer := grpc.NewServer()
	addr := ":8081"

	server := NewGRPCServer(grpcServer, addr)

	assert.Equal(t, addr, server.Address())
}

func TestGRPCServer_Stop(t *testing.T) {
	grpcServer := grpc.NewServer()
	addr := ":8081"

	server := NewGRPCServer(grpcServer, addr)

	// Stop should not return an error
	err := server.Stop(context.Background())
	assert.NoError(t, err)
}
