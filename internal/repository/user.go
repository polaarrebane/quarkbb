package repository

import (
	"context"
	"fmt"

	"codeberg.org/ronia/quarkbb/internal/model"
	"codeberg.org/ronia/quarkbb/internal/repository/sqlc"
)

// UserRepositoryImplementation is a concrete implementation of the UserRepository interface.
// It uses SQLC for type-safe database operations.
type UserRepositoryImplementation struct {
	q *sqlc.Queries
}

// NewUserRepository creates a new instance of the user repository implementation.
func NewUserRepository(q *sqlc.Queries) *UserRepositoryImplementation {
	return &UserRepositoryImplementation{
		q: q,
	}
}

// CreateUser persists a new user in the database.
func (ur UserRepositoryImplementation) CreateUser(ctx context.Context, user model.RegisterUserCommand) (*model.User, error) {
	cmd := sqlc.CreateUserAndReturnIdParams{
		Username: user.Username,
		Password: user.Password,
		Email:    user.Email,
	}
	id, err := ur.q.CreateUserAndReturnId(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	result := &model.User{
		ID:       id,
		Username: cmd.Username,
		Email:    cmd.Email,
	}

	return result, nil
}

// UsernameExists checks if a user with the given username already exists in the database.
func (ur UserRepositoryImplementation) UsernameExists(ctx context.Context, username string) (bool, error) {
	count, err := ur.q.CountUserByUsername(ctx, username)
	if err != nil {
		return false, fmt.Errorf("count user by username: %w", err)
	}

	return count > 0, nil
}
