package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	grpcContext "github.com/dtroode/urlshorter/internal/api/grpc/context"
	httpContext "github.com/dtroode/urlshorter/internal/api/http/context"
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
// It handles user authentication for both HTTP and gRPC transports.
type Authenticate struct {
	token              Token
	logger             *logger.Logger
	grpcContextManager *grpcContext.Manager
	httpContextManager *httpContext.Manager
}

// NewAuthenticate creates a new Authenticate middleware instance.
// It initializes authentication middleware with the provided token service and logger.
//
// Parameters:
//   - token: The token service for authentication operations
//   - l: The logger for authentication events
//
// Returns a pointer to the newly created Authenticate middleware instance.
func NewAuthenticate(token Token, l *logger.Logger) *Authenticate {
	return &Authenticate{
		token:              token,
		logger:             l,
		grpcContextManager: grpcContext.NewManager(),
		httpContextManager: httpContext.NewManager(),
	}
}

// authenticateUser contains the core, transport-agnostic authentication logic.
// It validates a token string and determines the user's identity.
// If the token is empty or invalid, it generates a new user ID.
// It returns the user ID, a flag indicating if a new token should be issued, and an error if authentication fails definitively.
func (m *Authenticate) authenticateUser(ctx context.Context, tokenString string) (userID uuid.UUID, issueNewToken bool, err error) {
	if tokenString == "" {
		// No token provided, treat as a new user.
		return uuid.New(), true, nil
	}

	userID, err = m.token.GetUserID(ctx, tokenString)
	if err != nil {
		// Any error from token validation is treated as a need for a new session.
		m.logger.Info("invalid token, creating new user session", "error", err)
		return uuid.New(), true, nil
	}

	if userID == uuid.Nil {
		// A nil UUID from a valid token is a specific failure case.
		return uuid.Nil, false, errors.New("authentication failed: invalid user id")
	}

	// Token is valid, user is authenticated.
	return userID, false, nil
}

// HandleHTTP is an adapter for http.Handler that uses the core authentication logic.
// It extracts the token from cookies, validates it, and sets the user ID in the request context.
// If authentication fails or a new token is needed, it sets appropriate cookies.
//
// Parameters:
//   - next: The next HTTP handler in the middleware chain
//
// Returns an HTTP handler that performs authentication before calling the next handler.
func (m *Authenticate) HandleHTTP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenString string
		cookie, err := r.Cookie("token")
		if err != nil && !errors.Is(err, http.ErrNoCookie) {
			m.logger.Error("failed to get cookie", "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		if cookie != nil {
			tokenString = cookie.Value
		}

		ctx := r.Context()
		userID, issueNewToken, authErr := m.authenticateUser(ctx, tokenString)
		if authErr != nil {
			http.Error(w, authErr.Error(), http.StatusUnauthorized)
			return
		}

		if issueNewToken {
			if err := m.setCookie(ctx, w, userID); err != nil {
				m.logger.Error("failed to set authentication cookie", "error", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
		}

		newCtx := m.httpContextManager.SetUserIDToContext(ctx, userID)
		next.ServeHTTP(w, r.WithContext(newCtx))
	})
}

// HandleGRPC is an adapter for grpc.UnaryServerInterceptor that uses the core authentication logic.
// It extracts the token from gRPC metadata, validates it, and sets the user ID in the request context.
// If authentication fails or a new token is needed, it sets appropriate metadata.
// It also sets the user ID in the response metadata for client reference.
//
// Parameters:
//   - ctx: The gRPC context
//   - req: The gRPC request
//   - _: The gRPC server info (unused)
//   - handler: The next gRPC handler in the interceptor chain
//
// Returns the gRPC response and any error that occurred during processing.
func (m *Authenticate) HandleGRPC(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	var tokenString string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if authHeaders := md.Get("authorization"); len(authHeaders) > 0 {
			tokenString = strings.TrimPrefix(authHeaders[0], "Bearer ")
		}
	}

	userID, issueNewToken, authErr := m.authenticateUser(ctx, tokenString)
	if authErr != nil {
		return nil, status.Error(codes.Unauthenticated, authErr.Error())
	}

	if issueNewToken {
		// Note: This sends the new token back in the response metadata.
		// The client is responsible for extracting and using it for subsequent requests.
		var err error
		ctx, err = m.setMetadata(ctx, userID)
		if err != nil {
			return nil, err // setMetadata already returns a gRPC status error
		}
	}

	m.logger.Info("user id", "userID", userID)

	newCtx := m.grpcContextManager.SetUserIDToContext(ctx, userID)

	return handler(newCtx, req)
}

// setCookie creates a new token and sets it in an HTTP cookie.
func (m *Authenticate) setCookie(ctx context.Context, w http.ResponseWriter, userID uuid.UUID) error {
	tokenString, err := m.token.CreateToken(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to create token: %w", err)
	}

	cookie := &http.Cookie{
		Name:     "token",
		Value:    tokenString,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	}

	http.SetCookie(w, cookie)
	return nil
}

// setMetadata creates a new token and prepares it for gRPC outgoing metadata.
func (m *Authenticate) setMetadata(ctx context.Context, userID uuid.UUID) (context.Context, error) {
	token, err := m.token.CreateToken(ctx, userID)
	if err != nil {
		m.logger.Error("failed to create token for grpc metadata", "error", err)
		return nil, status.Error(codes.Internal, "failed to create authentication token")
	}

	// The token is attached to the *outgoing* context of this specific RPC call.
	outCtx := metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
	err = grpc.SetHeader(outCtx, metadata.New(map[string]string{"authorization": "Bearer " + token}))
	if err != nil {
		m.logger.Error("failed to set header", "error", err)
		return nil, status.Error(codes.Internal, "failed to set header")
	}
	return outCtx, nil
}
