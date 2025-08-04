package model

import (
	"context"
	"net"
)

// SecurityLayer defines the interface for network security operations.
// It provides methods for creating network listeners with different security configurations.
type SecurityLayer interface {
	Listen(protocol, addr string) (net.Listener, error)
}

// Server defines the interface for server operations.
// It provides methods for starting, stopping, and managing server instances.
type Server interface {
	Start(securityLayer SecurityLayer) error
	Stop(ctx context.Context) error
	Address() string
}
