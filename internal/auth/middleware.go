package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

func (h *handler) JwtAuthMiddleware(next http.Handler) http.Handler {
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

		ctx := context.WithValue(r.Context(), "user", claims)
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
