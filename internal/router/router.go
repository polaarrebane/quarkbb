// Package router constructs and returns the application HTTP router
package router

import (
	"net/http"

	"codeberg.org/ronia/quarkbb/internal/auth"
	"github.com/go-chi/chi/v5"
)

// NewRouter is a constructor for app router
func NewRouter(auth *auth.Handler) http.Handler {
	r := chi.NewRouter()
	r.Post("/api/v1/register", auth.Register)
	return r
}
