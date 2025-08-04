package context

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey int

const (
	// userIDKey is the context key for storing user ID in request context.
	userIDKey contextKey = iota + 1
)

// Manager represents an HTTP context manager for user ID operations.
// It provides methods to set and retrieve user IDs from HTTP request context.
type Manager struct{}

// NewManager creates a new HTTP context manager instance.
//
// Returns a pointer to the newly created Manager instance.
func NewManager() *Manager {
	return &Manager{}
}

// SetUserIDToContext stores a user ID in the request context.
// This function is used by middleware to make the user ID available to handlers.
//
// Parameters:
//   - ctx: The request context to store the user ID in
//   - userID: The UUID of the user to store
//
// Returns a new context with the user ID stored.
func (m *Manager) SetUserIDToContext(ctx context.Context, userID uuid.UUID) context.Context {
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
func (m *Manager) GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	u, ok := ctx.Value(userIDKey).(uuid.UUID)
	return u, ok
}

// SetUserIDToResponse sets the user ID in the HTTP response headers.
// This function is used by middleware to make the user ID available to clients.
//
// Parameters:
//   - w: The HTTP response writer
//   - userID: The UUID of the user to set in the response
func (m *Manager) SetUserIDToResponse(w http.ResponseWriter, userID uuid.UUID) {
	w.Header().Set("X-User-ID", userID.String())
}

// GetUserIDFromResponse retrieves a user ID from the HTTP response headers.
// This function is used by clients to access the user ID that was set in the response.
//
// Parameters:
//   - resp: The HTTP response containing the user ID header
//
// Returns the user UUID and a boolean indicating if the user ID was found.
// If the user ID is not found, returns uuid.Nil and false.
func (m *Manager) GetUserIDFromResponse(resp *http.Response) (uuid.UUID, bool) {
	userIDStr := resp.Header.Get("X-User-ID")
	if userIDStr == "" {
		return uuid.Nil, false
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, false
	}

	return userID, true
}
