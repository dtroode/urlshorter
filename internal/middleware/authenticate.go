package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/dtroode/urlshorter/internal/auth"
	"github.com/dtroode/urlshorter/internal/logger"
)

// Token defines the interface for token-based authentication operations.
// It provides methods for creating and validating user tokens.
type Token interface {
	// GetUserID retrieves the user ID from a token string.
	// Returns the user UUID or an error if the token is invalid.
	GetUserID(ctx context.Context, tokenString string) (uuid.UUID, error)

	// CreateToken creates a new token for the specified user ID.
	// Returns the token string or an error if token creation fails.
	CreateToken(ctx context.Context, userID uuid.UUID) (string, error)
}

// Authenticate represents the authentication middleware.
// It handles user authentication via cookies and tokens.
type Authenticate struct {
	token  Token
	logger *logger.Logger
}

// NewAuthenticate creates a new Authenticate middleware instance.
//
// Parameters:
//   - token: The token service for authentication operations
//   - l: The logger instance for error logging
//
// Returns a pointer to the newly created Authenticate instance.
func NewAuthenticate(token Token, l *logger.Logger) *Authenticate {
	return &Authenticate{
		token:  token,
		logger: l,
	}
}

// Handle implements the http.Handler interface for authentication middleware.
// It processes incoming requests and handles user authentication via cookies.
// If no valid token is found, it creates a new user session automatically.
//
// The middleware:
//   - Extracts the token from the "token" cookie
//   - Validates the token and retrieves the user ID
//   - Creates a new user session if no valid token is found
//   - Sets the user ID in the request context for downstream handlers
//   - Sets a new authentication cookie if needed
func (m *Authenticate) Handle(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		cookie, err := r.Cookie("token")
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				userID := uuid.New()
				m.setCookie(ctx, w, userID)
				ctx := auth.SetUserIDToContext(r.Context(), userID)
				h.ServeHTTP(w, r.WithContext(ctx))
			} else {
				m.logger.Error("failed to get cookie", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
			}

			return
		}

		tokenString := cookie.Value

		userID, err := m.token.GetUserID(ctx, tokenString)
		if err != nil {
			userID := uuid.New()
			m.setCookie(ctx, w, userID)
			ctx := auth.SetUserIDToContext(r.Context(), userID)
			h.ServeHTTP(w, r.WithContext(ctx))

			return
		}

		if userID == uuid.Nil {
			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		ctx = auth.SetUserIDToContext(r.Context(), userID)

		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

// setCookie creates and sets an authentication cookie for the specified user.
// It generates a new token and sets it as an HTTP-only cookie with secure settings.
//
// Parameters:
//   - ctx: The request context
//   - w: The HTTP response writer
//   - userID: The user ID to create a token for
func (m *Authenticate) setCookie(ctx context.Context, w http.ResponseWriter, userID uuid.UUID) {
	tokenString, err := m.token.CreateToken(ctx, userID)
	if err != nil {
		m.logger.Error("failed to create token", "error", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	cookie := http.Cookie{
		Name:     "token",
		Value:    tokenString,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(w, &cookie)
}
