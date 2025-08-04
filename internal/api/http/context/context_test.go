package context

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_SetUserIDToContext(t *testing.T) {
	manager := NewManager()
	userID := uuid.New()
	ctx := context.Background()

	newCtx := manager.SetUserIDToContext(ctx, userID)

	// Check that the context contains the userID
	retrievedUserID, ok := manager.GetUserIDFromContext(newCtx)
	require.True(t, ok)
	assert.Equal(t, userID, retrievedUserID)
}

func TestManager_GetUserIDFromContext(t *testing.T) {
	manager := NewManager()
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		ctx := manager.SetUserIDToContext(context.Background(), userID)

		retrievedUserID, ok := manager.GetUserIDFromContext(ctx)
		require.True(t, ok)
		assert.Equal(t, userID, retrievedUserID)
	})

	t.Run("no user id in context", func(t *testing.T) {
		ctx := context.Background()

		retrievedUserID, ok := manager.GetUserIDFromContext(ctx)
		require.False(t, ok)
		assert.Equal(t, uuid.Nil, retrievedUserID)
	})

	t.Run("wrong type in context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), userIDKey, "not-a-uuid")

		retrievedUserID, ok := manager.GetUserIDFromContext(ctx)
		require.False(t, ok)
		assert.Equal(t, uuid.Nil, retrievedUserID)
	})
}
