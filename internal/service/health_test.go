package service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/dtroode/urlshorter/internal/service/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHealth_Ping(t *testing.T) {
	tests := map[string]struct {
		database      Database
		expectedError error
	}{
		"database is nil": {
			expectedError: nil,
		},
		"error from database": {
			database: (func() Database {
				databaseMock := mocks.NewDatabase(t)
				databaseMock.On("Ping", mock.Anything).Once().Return(errors.New("database error"))
				return databaseMock
			})(),
			expectedError: fmt.Errorf("failed to ping database: %w", errors.New("database error")),
		},
		"nil from database": {
			database: (func() Database {
				databaseMock := mocks.NewDatabase(t)
				databaseMock.On("Ping", mock.Anything).Once().Return(nil)
				return databaseMock
			})(),
			expectedError: nil,
		},
	}

	for tn, tt := range tests {
		tt := tt
		t.Run(tn, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			s := Health{
				DB: tt.database,
			}

			err := s.Ping(ctx)

			require.Equal(t, tt.expectedError, err)
		})
	}
}
