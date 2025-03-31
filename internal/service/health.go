package service

import (
	"context"
	"fmt"
)

type Database interface {
	Ping(ctx context.Context) error
}

type Health struct {
	DB Database
}

func (s *Health) Ping(ctx context.Context) error {
	if s.DB != nil {
		if err := s.DB.Ping(ctx); err != nil {
			return fmt.Errorf("failed to ping database: %w", err)
		}
	}

	return nil
}
