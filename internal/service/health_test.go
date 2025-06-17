package service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/dtroode/urlshorter/internal/service/mocks"
)

func TestHealth_Ping(t *testing.T) {
	tests := map[string]struct {
		pingers       []Pinger
		expectedError error
	}{
		"database is nil": {
			expectedError: nil,
		},
		"error from pinger": {
			pingers: []Pinger{(func() Pinger {
				pingerMock := mocks.NewPinger(t)
				pingerMock.On("Ping", mock.Anything).Once().Return(errors.New("database error"))
				return pingerMock
			})()},
			expectedError: fmt.Errorf("failed to ping service *mocks.Pinger"),
		},
		"nil from pinger": {
			pingers: []Pinger{(func() Pinger {
				pingerMock := mocks.NewPinger(t)
				pingerMock.On("Ping", mock.Anything).Once().Return(nil)
				return pingerMock
			})()},
			expectedError: nil,
		},
	}

	for tn, tt := range tests {
		tt := tt
		t.Run(tn, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			s := Health{
				pingers: tt.pingers,
			}

			err := s.Ping(ctx)

			require.Equal(t, tt.expectedError, err)
		})
	}
}
