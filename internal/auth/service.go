package auth

import (
	"context"
	"fmt"
	"time"

	"codeberg.org/ronia/quarkbb/internal/model"
	r "codeberg.org/ronia/quarkbb/internal/repository"
	"codeberg.org/ronia/quarkbb/internal/security"
)

// registeredUser represents the response data after successful user registration.
// It contains the user ID and username for the newly created account.
type registeredUser struct {
	ID       int64  `json:"id"`       // Unique identifier for the user
	Username string `json:"username"` // User's chosen username
}

// signedInUser represents user information returned after successful login.
// It contains the user ID and username for authentication response.
type signedInUser struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

// accessToken represents the JWT access token response.
// It contains the token string, expiration time, and user information.
type accessToken struct {
	RawString string       `json:"access_token"`
	ExpiresIn time.Time    `json:"expires_in"`
	User      signedInUser `json:"user"`
}

// refreshToken represents the JWT refresh token response.
// It contains only the raw token string for client-side storage.
type refreshToken struct {
	RawString string `json:"refresh_token"`
}

type authServiceImpl struct {
	r  r.UserRepository // Repository for database operations
	js security.JWTService
}

// Service provides authentication and user management business logic.
// It acts as the intermediate layer between HTTP handlers and the repository layer,
// implementing validation, security checks, and business rules.
type AuthService interface {
	Register(ctx context.Context, c model.RegisterUserCommand) (*registeredUser, error)
	Login(ctx context.Context, c model.LoginCommand) (*accessToken, *refreshToken, error)
	VerifyAuthToken(tokenString string) (*security.AuthClaims, error)
}

// NewService creates a new instance of the authentication service.
// It requires a UserRepository implementation for data access operations.
// Returns a pointer to the initialized Service.
func NewService(r r.UserRepository, js security.JWTService) AuthService {
	return &authServiceImpl{
		r:  r,
		js: js,
	}
}

// VerifyAuthToken validates an access token string and returns the claims.
// It verifies the signature, expiration, and other JWT claims.
// Returns an error if the token is invalid or expired.
func (svc authServiceImpl) VerifyAuthToken(tokenString string) (*security.AuthClaims, error) {
	return svc.js.VerifyAuthToken(tokenString)
}

// Register creates a new user account in the system.
// It validates that the username doesn't exist, hashes the password,
// and persists the user data to the database.
// Returns the registered user information or an appropriate error.
func (svc authServiceImpl) Register(ctx context.Context, c model.RegisterUserCommand) (*registeredUser, error) {
	exists, err := svc.r.UsernameExists(ctx, c.Username)
	if err != nil {
		// todo: add log
		return nil, errRegistrationFailed
	}

	if exists {
		return nil, errUserAlreadyExists
	}

	hash, err := security.GeneratePasswordHash(c.Password)
	if err != nil {
		// todo: add log
		return nil, errRegistrationFailed
	}

	cmd := model.RegisterUserCommand{
		Username: c.Username,
		Email:    c.Email,
		Password: hash,
	}

	user, err := svc.r.CreateUser(ctx, cmd)
	if err != nil {
		// todo: add log
		return nil, errRegistrationFailed
	}

	result := &registeredUser{
		ID:       user.ID,
		Username: user.Username,
	}

	return result, nil
}

// Login authenticates a user with username and password.
// It validates credentials and generates both access and refresh tokens.
// Returns token pair on success or appropriate authentication error.
func (svc authServiceImpl) Login(ctx context.Context, c model.LoginCommand) (*accessToken, *refreshToken, error) {
	user, err := svc.r.FindUserByUsername(ctx, c.Username)
	if err != nil {
		return nil, nil, errUserNotFound
	}

	if !security.CheckPassword(c.Password, user.Password) {
		return nil, nil, errWrongPassword
	}

	username := user.Username
	userID := fmt.Sprintf("%d", user.ID)
	tokens, _ := svc.js.GenerateTokenPair(username, userID)
	at := &accessToken{
		RawString: tokens.AccessToken,
		ExpiresIn: tokens.AccessExpiresIn,
		User: signedInUser{
			Username: user.Username,
			ID:       user.ID,
		},
	}
	rt := &refreshToken{
		RawString: tokens.RefreshToken,
	}
	return at, rt, nil
}
