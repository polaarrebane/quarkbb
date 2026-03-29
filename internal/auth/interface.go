package auth

import (
	"context"
	"net/http"

	"codeberg.org/ronia/quarkbb/internal/model"
	"codeberg.org/ronia/quarkbb/internal/security"
)

const refreshTokenCookieName = "refresh_token"
const csrfTokenCookieName = "csrf_token"

// Handler processes HTTP requests for authentication endpoints.
type Handler interface {
	Register(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
	Sessions(w http.ResponseWriter, r *http.Request)
	Refresh(w http.ResponseWriter, r *http.Request)
	CloseSession(w http.ResponseWriter, r *http.Request)

	JwtAuthTokenMiddleware(next http.Handler) http.Handler
	JwtRefreshTokenMiddleware(next http.Handler) http.Handler
	SessionIDCtx(next http.Handler) http.Handler
}

// Service provides authentication and user management business logic.
// It acts as the intermediate layer between HTTP handlers and the repository layer,
// implementing validation, security checks, and business rules.
type Service interface {
	Register(ctx context.Context, c model.RegisterUserCommand) (*registeredUser, error)
	Login(ctx context.Context, c model.LoginCommand) (*accessToken, *refreshToken, error)
	Refresh(ctx context.Context, c model.RefreshCommand) (*accessToken, *refreshToken, error)
	Logout(ctx context.Context, c model.LogoutCommand) error
	Sessions(ctx context.Context, cmd model.RetrieveSessionsCommand) ([]model.Session, error)
	CloseSession(ctx context.Context, cmd model.CloseSessionCommand) error
	VerifyAuthToken(tokenString string) (*security.AuthClaims, error)
	VerifyRefreshToken(tokenString string) (*security.RefreshClaims, error)
	VerifyCsrfToken(csrfToken string, jti string) bool
	GenerateCsrfToken(jti string) string
}
