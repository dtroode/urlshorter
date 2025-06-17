package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/dtroode/urlshorter/internal/auth"
	"github.com/dtroode/urlshorter/internal/logger"
)

type Token interface {
	GetUserID(ctx context.Context, tokenString string) (uuid.UUID, error)
	CreateToken(ctx context.Context, userID uuid.UUID) (string, error)
}

type Authenticate struct {
	token  Token
	logger *logger.Logger
}

func NewAuthenticate(token Token, l *logger.Logger) *Authenticate {
	return &Authenticate{
		token:  token,
		logger: l,
	}
}

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
