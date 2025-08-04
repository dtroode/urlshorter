package context

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func TestManager_SetUserIDToContext(t *testing.T) {
	manager := NewManager()
	userID := uuid.New()
	ctx := context.Background()

	newCtx := manager.SetUserIDToContext(ctx, userID)

	// Check that metadata was added to incoming context
	md, ok := metadata.FromIncomingContext(newCtx)
	require.True(t, ok)

	// Verify that the user ID is in the metadata
	userIDs := md.Get("user_id")
	require.Len(t, userIDs, 1)
	assert.Equal(t, userID.String(), userIDs[0])
}

func TestManager_GetUserIDFromContext(t *testing.T) {
	manager := NewManager()
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		// Create context with user ID using the public API
		ctx := manager.SetUserIDToContext(context.Background(), userID)

		retrievedUserID, ok := manager.GetUserIDFromContext(ctx)
		require.True(t, ok)
		assert.Equal(t, userID, retrievedUserID)
	})

	t.Run("no metadata", func(t *testing.T) {
		ctx := context.Background()

		retrievedUserID, ok := manager.GetUserIDFromContext(ctx)
		require.False(t, ok)
		assert.Equal(t, uuid.Nil, retrievedUserID)
	})

	t.Run("no user_id in metadata", func(t *testing.T) {
		md := metadata.New(map[string]string{"other_key": "other_value"})
		ctx := metadata.NewIncomingContext(context.Background(), md)

		retrievedUserID, ok := manager.GetUserIDFromContext(ctx)
		require.False(t, ok)
		assert.Equal(t, uuid.Nil, retrievedUserID)
	})

	t.Run("invalid uuid", func(t *testing.T) {
		// Create metadata with invalid UUID using the known key
		md := metadata.New(map[string]string{"user_id": "invalid-uuid"})
		ctx := metadata.NewIncomingContext(context.Background(), md)

		retrievedUserID, ok := manager.GetUserIDFromContext(ctx)
		require.False(t, ok)
		assert.Equal(t, uuid.Nil, retrievedUserID)
	})
}
