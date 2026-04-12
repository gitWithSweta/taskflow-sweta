package middleware

import (
	"context"
	"net/http"
	"strings"

	"taskflow/internal/auth"

	"github.com/google/uuid"
)

type ctxKey int

const (
	userIDKey    ctxKey = 1
	sessionIDKey ctxKey = 2
)

type SessionValidator interface {
	Exists(ctx context.Context, sessionID uuid.UUID) (bool, error)
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	v := ctx.Value(userIDKey)
	if v == nil {
		return uuid.Nil, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok
}

func SessionIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	v := ctx.Value(sessionIDKey)
	if v == nil {
		return uuid.Nil, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok
}

func RequireAuth(secret []byte, sessions SessionValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if h == "" {
				writeUnauthorized(w)
				return
			}
			const prefix = "Bearer "
			if len(h) < len(prefix) || !strings.EqualFold(h[:len(prefix)], prefix) {
				writeUnauthorized(w)
				return
			}
			raw := strings.TrimSpace(h[len(prefix):])
			if raw == "" {
				writeUnauthorized(w)
				return
			}

			claims, err := auth.ParseToken(secret, raw)
			if err != nil {
				writeUnauthorized(w)
				return
			}
			uid, err := uuid.Parse(claims.UserID)
			if err != nil {
				writeUnauthorized(w)
				return
			}
			sessionID, err := claims.SessionID()
			if err != nil {
				writeUnauthorized(w)
				return
			}

			ok, err := sessions.Exists(r.Context(), sessionID)
			if err != nil || !ok {
				writeUnauthorized(w)
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, userIDKey, uid)
			ctx = context.WithValue(ctx, sessionIDKey, sessionID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
}
