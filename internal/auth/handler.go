// Package auth handles user registration and authentication HTTP requests.
// It provides HTTP handlers for processing authentication-related endpoints,
// request validation, and response formatting.
package auth

import (
	"errors"
	"net/http"

	"codeberg.org/ronia/quarkbb/internal/model"
)

// Handler processes HTTP requests for authentication endpoints.
type Handler struct {
	svc Service
}

// NewAuthHandler creates a new authentication HTTP handler.
func NewAuthHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// Register handles POST /api/v1/register requests.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	limitBodySize(w, r)

	var cmd model.RegisterUserCommand
	if resp, err := decodeAndValidateCommand(r, &cmd); err != nil {
		resp.Send(w)
		return
	}

	data, err := h.svc.register(r.Context(), cmd)
	if err != nil {
		switch {
		case errors.Is(err, errUserAlreadyExists):
			newUserAlreadyExistsResponse().Send(w)
			return
		}
		newRegistrationFailedResponse(errRegistrationFailed).Send(w)
		return
	}

	newCreatedResponse(data).Send(w)
}
