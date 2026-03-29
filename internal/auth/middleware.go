package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

type contextKey string

const (
	authClaimsContextKey    contextKey = "authClaimsContextKey"
	refreshClaimsContextKey contextKey = "refreshClaimsContextKey"
	sessionIDContextKey     contextKey = "sessionIDContextKey"
)

func (h *handler) JwtAuthTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, err := extractBearerToken(r)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		claims, err := h.svc.VerifyAuthToken(raw)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), authClaimsContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *handler) JwtRefreshTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenCookie, err := r.Cookie(refreshTokenCookieName)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		claims, err := h.svc.VerifyRefreshToken(tokenCookie.Value)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		ctx := context.WithValue(r.Context(), refreshClaimsContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *handler) SessionIDCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionID := chi.URLParam(r, "sessionID")
		ctx := context.WithValue(r.Context(), sessionIDContextKey, sessionID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func extractBearerToken(r *http.Request) (string, error) {
	prefix := "Bearer "
	h := r.Header.Get("Authorization")
	token, found := strings.CutPrefix(h, prefix)
	if !found {
		return "", errors.New("prefix `Bearer` not found")
	}
	return token, nil
}
