package auth

import (
	"context"
	"errors"
	"fmt"

	"codeberg.org/ronia/quarkbb/internal/model"
	"codeberg.org/ronia/quarkbb/internal/repository"
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
	ExpiresIn int          `json:"expires_in"`
	User      signedInUser `json:"user"`
}

// refreshToken represents the JWT refresh token response.
// It contains only the raw token string for client-side storage.
type refreshToken struct {
	RawString string `json:"refresh_token"`
	MaxAge    int    `json:"expires_in"`
}

type serviceImpl struct {
	users  repository.UserRepository
	tokens repository.RefreshTokenRepository
	jwt    security.JWTService
}

// Service provides authentication and user management business logic.
// It acts as the intermediate layer between HTTP handlers and the repository layer,
// implementing validation, security checks, and business rules.
type Service interface {
	Register(ctx context.Context, c model.RegisterUserCommand) (*registeredUser, error)
	Login(ctx context.Context, c model.LoginCommand) (*accessToken, *refreshToken, error)
	Refresh(ctx context.Context, c model.RefreshCommand) (*accessToken, *refreshToken, error)
	VerifyAuthToken(tokenString string) (*security.AuthClaims, error)
	VerifyRefreshToken(tokenString string) (*security.RefreshClaims, error)
}

// NewService creates a new instance of the authentication service.
func NewService(
	ru repository.UserRepository,
	rt repository.RefreshTokenRepository,
	js security.JWTService,
) Service {
	return &serviceImpl{
		users:  ru,
		tokens: rt,
		jwt:    js,
	}
}

// VerifyAuthToken validates an access token string and returns the claims.
// It verifies the signature, expiration, and other JWT claims.
// Returns an error if the token is invalid or expired.
func (svc serviceImpl) VerifyAuthToken(tokenString string) (*security.AuthClaims, error) {
	claims, err := svc.jwt.VerifyAuthToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("verify auth token: %w", err)
	}
	return claims, nil
}

// VerifyRefreshToken validates a refresh token string and returns the claims.
func (svc serviceImpl) VerifyRefreshToken(tokenString string) (*security.RefreshClaims, error) {
	claims, err := svc.jwt.VerifyRefreshToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("verify refresh token: %w", err)
	}
	return claims, nil
}

// Register creates a new user account in the system.
// It validates that the username doesn't exist, hashes the password,
// and persists the user data to the database.
// Returns the registered user information or an appropriate error.
func (svc serviceImpl) Register(ctx context.Context, c model.RegisterUserCommand) (*registeredUser, error) {
	exists, err := svc.users.UsernameExists(ctx, c.Username)
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

	user, err := svc.users.CreateUser(ctx, cmd)
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
func (svc serviceImpl) Login(ctx context.Context, c model.LoginCommand) (*accessToken, *refreshToken, error) {
	user, err := svc.users.FindUserByUsername(ctx, c.Username)
	if err != nil {
		var nfe *repository.NotFoundError
		if errors.As(err, &nfe) {
			return nil, nil, errUserNotFound
		}
		return nil, nil, errLoginFailed
	}

	if !security.CheckPassword(c.Password, user.Password) {
		return nil, nil, errWrongPassword
	}

	at, rt, err := svc.createTokenPair(user)
	if err != nil {
		return nil, nil, fmt.Errorf("create token pair: %w", err)
	}
	return at, rt, nil
}

// Refresh an auth token.
func (svc serviceImpl) Refresh(ctx context.Context, cmd model.RefreshCommand) (*accessToken, *refreshToken, error) {
	user, err := svc.users.GetUserByID(ctx, cmd.UserID)
	if err != nil {
		var nfe *repository.NotFoundError
		if errors.As(err, &nfe) {
			return nil, nil, errUserNotFound
		}
		return nil, nil, errRefreshFailed
	}

	used, err := svc.tokens.RefreshTokenUsed(ctx, cmd.JTI)
	if err != nil {
		return nil, nil, fmt.Errorf("database error: %w", err)
	}
	if used {
		// todo: revoke all sessions
		return nil, nil, errors.New("used token detected")
	}

	if err := svc.tokens.MarkTokenAsUsed(ctx, cmd.JTI); err != nil {
		return nil, nil, fmt.Errorf("database error: %w", err)
	}

	at, rt, err := svc.createTokenPair(user)
	if err != nil {
		return nil, nil, fmt.Errorf("refresh error: %w", err)
	}
	return at, rt, nil
}

func (svc serviceImpl) createTokenPair(user *model.User) (*accessToken, *refreshToken, error) {
	tokens, err := svc.jwt.GenerateTokenPair(user.Username, fmt.Sprintf("%d", user.ID))
	if err != nil {
		return nil, nil, fmt.Errorf("create token pair error: %w", err)
	}
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
		MaxAge:    tokens.RefreshExpiresIn,
	}
	return at, rt, nil
}
