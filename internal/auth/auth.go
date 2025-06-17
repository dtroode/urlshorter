package auth

import (
	"context"

	"github.com/google/uuid"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey int

const (
	// userIDKey is the context key for storing user ID in request context.
	userIDKey contextKey = iota + 1
)

// SetUserIDToContext stores a user ID in the request context.
// This function is used by middleware to make the user ID available to handlers.
//
// Parameters:
//   - ctx: The request context to store the user ID in
//   - userID: The UUID of the user to store
//
// Returns a new context with the user ID stored.
func SetUserIDToContext(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// GetUserIDFromContext retrieves a user ID from the request context.
// This function is used by handlers to access the authenticated user's ID.
//
// Parameters:
//   - ctx: The request context containing the user ID
//
// Returns the user UUID and a boolean indicating if the user ID was found.
// If the user ID is not found, returns uuid.Nil and false.
func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	u, ok := ctx.Value(userIDKey).(uuid.UUID)
	return u, ok
}
