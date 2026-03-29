// Package router constructs and returns the application HTTP router
package router

import (
	"encoding/json"
	"net/http"

	"codeberg.org/ronia/quarkbb/internal/auth"
	"github.com/go-chi/chi/v5"
)

// NewRouter creates and configures the HTTP router with authentication middleware.
func NewRouter(auth auth.Handler) http.Handler {
	r := chi.NewRouter()
	r.Post("/api/v1/auth/login", auth.Login)
	r.Post("/api/v1/auth/register", auth.Register)

	r.With(auth.JwtRefreshTokenMiddleware).Post("/api/v1/auth/refresh", auth.Refresh)
	r.With(auth.JwtRefreshTokenMiddleware).Post("/api/v1/auth/logout", auth.Logout)

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(auth.JwtAuthTokenMiddleware)
		r.Get("/auth/sessions", auth.Sessions)
		r.With(auth.SessionIDCtx).Delete("/auth/sessions/{sessionID}", auth.CloseSession)
		r.Get("/protected", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode("it works"); err != nil {
				http.Error(w, "failed to encode response", http.StatusInternalServerError)
			}
		})
	})

	return r
}
