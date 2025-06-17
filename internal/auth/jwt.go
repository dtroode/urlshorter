package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims represents the JWT claims structure.
// It extends jwt.RegisteredClaims with a custom UserID field.
type Claims struct {
	jwt.RegisteredClaims
	// UserID is the UUID of the user associated with this token.
	UserID uuid.UUID `json:"user_id"`
}

// JWT represents a JWT token service for authentication.
// It provides methods for creating and validating JWT tokens.
type JWT struct {
	// secretKey is the secret key used for signing and validating JWT tokens.
	secretKey string
}

// NewJWT creates a new JWT service instance with the provided secret key.
//
// Parameters:
//   - secretKey: The secret key used for JWT signing and validation
//
// Returns a pointer to the newly created JWT instance.
func NewJWT(secretKey string) *JWT {
	return &JWT{
		secretKey: secretKey,
	}
}

// GetUserID extracts and validates a user ID from a JWT token string.
// It parses the token, validates its signature, and extracts the user ID from claims.
//
// Parameters:
//   - ctx: The request context (unused in current implementation)
//   - tokenString: The JWT token string to parse and validate
//
// Returns the user UUID from the token claims or an error if validation fails.
// If the token is invalid or expired, returns uuid.Nil and an error.
func (j *JWT) GetUserID(_ context.Context, tokenString string) (uuid.UUID, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("wrong signing method %v", t.Header["alg"])
		}

		return []byte(j.secretKey), nil
	})

	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return uuid.Nil, fmt.Errorf("token is invalid")
	}

	return claims.UserID, nil
}

// CreateToken creates a new JWT token for the specified user ID.
// The token includes standard claims (issued at, expires at) and the user ID.
// Tokens are valid for 24 hours from creation.
//
// Parameters:
//   - ctx: The request context (unused in current implementation)
//   - userID: The UUID of the user to create a token for
//
// Returns the signed JWT token string or an error if token creation fails.
func (j *JWT) CreateToken(_ context.Context, userID uuid.UUID) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(j.secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}
