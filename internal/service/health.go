package service

import (
	"context"
	"fmt"
	"reflect"
)

// Pinger defines interface for services that can be pinged.
type Pinger interface {
	Ping(ctx context.Context) error
}

// Health represents health check service.
type Health struct {
	pingers []Pinger
}

// NewHealth creates new Health instance.
func NewHealth(pingers ...Pinger) *Health {
	return &Health{
		pingers: pingers,
	}
}

// Ping checks health of all registered services.
func (s *Health) Ping(ctx context.Context) error {
	for _, pinger := range s.pingers {
		if err := pinger.Ping(ctx); err != nil {
			return fmt.Errorf("failed to ping service %s", reflect.TypeOf(pinger))
		}
	}

	return nil
}
