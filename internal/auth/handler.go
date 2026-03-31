// Package auth handles user registration and authentication HTTP requests.
// It provides HTTP handlers for processing authentication-related endpoints,
// request validation, and response formatting.
package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"codeberg.org/ronia/quarkbb/internal/model"
	"codeberg.org/ronia/quarkbb/internal/security"
	"codeberg.org/ronia/quarkbb/internal/validator"
)

type handler struct {
	svc       Service
	validator validator.Validator
	revoked   RevokedTokensStorage
}

// NewHandler creates a new authentication HTTP handler.
func NewHandler(svc Service, val validator.Validator, revoked RevokedTokensStorage) Handler {
	return &handler{
		svc:       svc,
		validator: val,
		revoked:   revoked,
	}
}

// Register handles POST /api/v1/auth/register requests.
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

// Login handles POST /api/v1/auth/login requests.
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
	h.setRefreshTokenCookie(w, refreshToken)
	h.setCsrfTokenCookie(w, refreshToken)
	newAccessGrantedResponse(accessToken).Send(w)
}

// Refresh handles POST /api/v1/auth/refresh requests.
func (h *handler) Refresh(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, err := extractRefreshClaims(ctx)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	sessionID := claims.SessionID
	jti := claims.ID
	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	at, rt, err := h.svc.Refresh(ctx, model.RefreshCommand{
		UserID:    userID,
		JTI:       jti,
		SessionID: sessionID,
	})
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	h.setRefreshTokenCookie(w, rt)
	h.setCsrfTokenCookie(w, rt)
	newAccessGrantedResponse(at).Send(w)
}

// Logout handles POST /api/v1/auth/logout requests.
func (h *handler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	refreshClaims, err := extractRefreshClaims(ctx)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	sessionID := refreshClaims.SessionID
	refreshJti := refreshClaims.ID
	userID, err := strconv.ParseInt(refreshClaims.Subject, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	authClaims, err := extractAuthClaims(ctx)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	authJti := authClaims.ID

	if err := h.svc.Logout(r.Context(), model.LogoutCommand{
		UserID:     userID,
		AuthJTI:    authJti,
		RefreshJTI: refreshJti,
		SessionID:  sessionID,
	}); err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	fmt.Printf("logout: %s\n", authJti)
	h.unsetRefreshTokenCookie(w)
	h.unsetCsrfTokenCookie(w)
}

// Sessions handles GET /api/v1/auth/sessions requests.
func (h *handler) Sessions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, err := extractAuthClaims(ctx)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	cmd := model.RetrieveSessionsCommand{
		Username: claims.Username,
	}

	sessions, err := h.svc.Sessions(ctx, cmd)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	currentSession := claims.SessionID
	newSessionsResponse(currentSession, sessions).Send(w)
}

// CloseSession handles DELETE /api/v1/auth/sessions/{sessionID} requests.
func (h *handler) CloseSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	authClaims, err := extractAuthClaims(ctx)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	sessionID, err := extractSessionID(ctx)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	cmd := model.CloseSessionCommand{
		CurrentSession: authClaims.SessionID,
		SessionID:      sessionID,
	}

	if err := h.svc.CloseSession(ctx, cmd); err != nil {
		switch {
		case errors.Is(err, errSessionNotFound):
			w.WriteHeader(http.StatusNotFound)
			return
		case errors.Is(err, errMalformedID):
			writeError(w, http.StatusBadRequest, err.Error(), nil)
			return
		case errors.Is(err, errCantCloseCurrentSession):
			writeError(w, http.StatusBadRequest, err.Error(), nil)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) decodeAndValidateCommand(r *http.Request, cmd any) (*response, error) {
	if err := h.decodeCommand(r, cmd); err != nil {
		return newMalformedBodyResponse(), err
	}

	locale := parseLocale(r.Header.Get("Accept-Language"))
	if err := h.validator.ValidateCommand(cmd, locale); err != nil {
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

func (h *handler) setRefreshTokenCookie(w http.ResponseWriter, rt *refreshToken) {
	cookieRefreshToken := http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    rt.RawString,
		Path:     "/api/v1/auth",
		MaxAge:   rt.MaxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, &cookieRefreshToken)
}

func (h *handler) setCsrfTokenCookie(w http.ResponseWriter, rt *refreshToken) {
	csrfToken := h.svc.GenerateCsrfToken(rt.ID)
	cookieRefreshToken := http.Cookie{
		Name:     csrfTokenCookieName,
		Value:    csrfToken,
		Path:     "/api/v1/auth",
		MaxAge:   rt.MaxAge,
		HttpOnly: false,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, &cookieRefreshToken)
}

func (h *handler) unsetRefreshTokenCookie(w http.ResponseWriter) {
	cookieRefreshToken := http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    "",
		Path:     "/api/v1/auth",
		MaxAge:   0,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, &cookieRefreshToken)
}

func (h *handler) unsetCsrfTokenCookie(w http.ResponseWriter) {
	cookieRefreshToken := http.Cookie{
		Name:     csrfTokenCookieName,
		Value:    "",
		Path:     "/api/v1/auth",
		MaxAge:   0,
		HttpOnly: false,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, &cookieRefreshToken)
}

func extractAuthClaims(ctx context.Context) (*security.AuthClaims, error) {
	authFromContext := ctx.Value(authClaimsContextKey)
	if authFromContext == nil {
		return nil, errAuthorizationRequired
	}

	authClaims, ok := authFromContext.(*security.AuthClaims)
	if !ok {
		return nil, errAuthorizationRequired
	}

	return authClaims, nil
}

func extractRefreshClaims(ctx context.Context) (*security.RefreshClaims, error) {
	refreshFromContext := ctx.Value(refreshClaimsContextKey)
	if refreshFromContext == nil {
		return nil, errAuthorizationRequired
	}

	refreshClaims, ok := refreshFromContext.(*security.RefreshClaims)
	if !ok {
		return nil, errAuthorizationRequired
	}

	return refreshClaims, nil
}

func extractSessionID(ctx context.Context) (string, error) {
	sessionIDFromContext := ctx.Value(sessionIDContextKey)
	if sessionIDFromContext == nil {
		return "", errSessionNotFound
	}
	sessionID, ok := sessionIDFromContext.(string)
	if !ok {
		return "", errSessionNotFound
	}
	return sessionID, nil
}
