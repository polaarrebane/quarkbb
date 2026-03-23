// Package auth handles user registration and authentication HTTP requests.
// It provides HTTP handlers for processing authentication-related endpoints,
// request validation, and response formatting.
package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"codeberg.org/ronia/quarkbb/internal/model"
	"codeberg.org/ronia/quarkbb/internal/validator"
)

type handler struct {
	svc Service
	val validator.Validator
}

// Handler processes HTTP requests for authentication endpoints.
type Handler interface {
	Register(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	JwtAuthMiddleware(next http.Handler) http.Handler
}

// NewHandler creates a new authentication HTTP handler.
func NewHandler(svc Service, val validator.Validator) Handler {
	return &handler{
		svc: svc,
		val: val,
	}
}

// Register handles POST /api/v1/register requests.
func (h *handler) Register(w http.ResponseWriter, r *http.Request) {
	limitBodySize(w, r)

	var cmd model.RegisterUserCommand
	if resp, err := h.decodeAndValidateCommand(r, &cmd); err != nil {
		resp.Send(w)
		return
	}

	data, err := h.svc.Register(r.Context(), cmd)
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

// Login handles POST /api/v1/login requests.
func (h *handler) Login(w http.ResponseWriter, r *http.Request) {
	limitBodySize(w, r)

	var cmd model.LoginCommand
	if resp, err := h.decodeAndValidateCommand(r, &cmd); err != nil {
		resp.Send(w)
		return
	}

	accessToken, refreshToken, err := h.svc.Login(r.Context(), cmd)
	if err != nil {
		switch {
		case errors.Is(err, errUserNotFound):
			newUserNotFoundResponse().Send(w)
			return
		case errors.Is(err, errWrongPassword):
			newWrongPasswordResponse().Send(w)
			return
		}
		newLoginFailedResponse().Send(w)
		return
	}

	cookieRefreshToken := http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken.RawString,
		Path:     "/auth",
		MaxAge:   refreshToken.MaxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, &cookieRefreshToken)
	newSignedInResponse(accessToken).Send(w)
}

func (h *handler) decodeAndValidateCommand(r *http.Request, cmd any) (*response, error) {
	if err := h.decodeCommand(r, cmd); err != nil {
		return newMalformedBodyResponse(), err
	}

	locale := parseLocale(r.Header.Get("Accept-Language"))
	if err := h.val.ValidateCommand(cmd, locale); err != nil {
		var ve *validator.ValidationErrors
		if errors.As(err, &ve) {
			return newValidationErrorResponse(ve), err
		}
		return newInternalErrorResponse(), err
	}

	return nil, nil
}

func (h *handler) decodeCommand(r *http.Request, cmd any) error {
	if err := json.NewDecoder(r.Body).Decode(cmd); err != nil {
		return fmt.Errorf("decode request body: %w", err)
	}
	return nil
}
