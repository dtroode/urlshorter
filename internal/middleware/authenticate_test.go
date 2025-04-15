package middleware

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dtroode/urlshorter/internal/auth"
	"github.com/dtroode/urlshorter/internal/logger"
	"github.com/dtroode/urlshorter/internal/middleware/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthenticate_Handle(t *testing.T) {
	dummyLogger := &logger.Logger{
		Logger: slog.New(slog.NewJSONHandler(io.Discard, nil)),
	}

	userID := uuid.New()
	tokenString := "token-string"

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contextUserID, ok := auth.GetUserIDFromContext(r.Context())
		assert.True(t, ok)
		assert.NotEqual(t, uuid.Nil, contextUserID)
		w.WriteHeader(http.StatusOK)
	})

	t.Run("no cookie", func(t *testing.T) {
		t.Parallel()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := r.Context()

		tokenMock := mocks.NewToken(t)
		tokenMock.On("Create", ctx, mock.AnythingOfType("uuid.UUID")).Once().
			Return(tokenString, nil)

		m := NewAuthenticate(tokenMock, dummyLogger)
		m.Handle(h).ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Equal(t, "token", w.Result().Cookies()[0].Name)
		assert.Equal(t, tokenString, w.Result().Cookies()[0].Value)
	})

	t.Run("failed to create token", func(t *testing.T) {
		t.Parallel()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := r.Context()

		tokenMock := mocks.NewToken(t)
		tokenMock.On("Create", ctx, mock.AnythingOfType("uuid.UUID")).Once().
			Return("", errors.New("token error"))

		m := NewAuthenticate(tokenMock, dummyLogger)
		m.Handle(h).ServeHTTP(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
		assert.Len(t, w.Result().Cookies(), 0)
	})

	t.Run("failed to get user id", func(t *testing.T) {
		t.Parallel()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := r.Context()

		cookie := http.Cookie{
			Name:     "token",
			Value:    "existing-token",
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		}
		r.AddCookie(&cookie)

		tokenMock := mocks.NewToken(t)
		tokenMock.On("GetUserID", ctx, "existing-token").Once().
			Return(uuid.Nil, errors.New("token error"))
		tokenMock.On("Create", ctx, mock.AnythingOfType("uuid.UUID")).Once().
			Return(tokenString, nil)

		m := NewAuthenticate(tokenMock, dummyLogger)
		m.Handle(h).ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Equal(t, "token", w.Result().Cookies()[0].Name)
		assert.Equal(t, tokenString, w.Result().Cookies()[0].Value)
	})

	t.Run("user id is nil", func(t *testing.T) {
		t.Parallel()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := r.Context()

		cookie := http.Cookie{
			Name:     "token",
			Value:    "existing-token",
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		}
		r.AddCookie(&cookie)

		tokenMock := mocks.NewToken(t)
		tokenMock.On("GetUserID", ctx, "existing-token").Once().
			Return(uuid.Nil, nil)

		m := NewAuthenticate(tokenMock, dummyLogger)
		m.Handle(h).ServeHTTP(w, r)

		assert.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)
		assert.Len(t, w.Result().Cookies(), 0)
	})

	t.Run("user id is valid", func(t *testing.T) {
		t.Parallel()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := r.Context()

		cookie := http.Cookie{
			Name:     "token",
			Value:    "existing-token",
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		}
		r.AddCookie(&cookie)

		tokenMock := mocks.NewToken(t)
		tokenMock.On("GetUserID", ctx, "existing-token").Once().
			Return(userID, nil)

		m := NewAuthenticate(tokenMock, dummyLogger)
		m.Handle(h).ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	})
}
