package auth

import (
	"context"

	"codeberg.org/ronia/quarkbb/internal/model"
	r "codeberg.org/ronia/quarkbb/internal/repository"
	"codeberg.org/ronia/quarkbb/internal/security"
)

type registeredUser struct {
	ID       int64  `json:"id"`       // Unique identifier for the user
	Username string `json:"username"` // User's chosen username
}

// Service provides authentication and user management business logic.
// It acts as the intermediate layer between HTTP handlers and the repository layer,
// implementing validation, security checks, and business rules.
type Service struct {
	r r.UserRepository // Repository for database operations
}

// NewService creates a new instance of the authentication service.
// It requires a UserRepository implementation for data access operations.
// Returns a pointer to the initialized Service.
func NewService(r r.UserRepository) *Service {
	return &Service{r: r}
}

func (svc Service) register(ctx context.Context, c model.RegisterUserCommand) (*registeredUser, error) {
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
