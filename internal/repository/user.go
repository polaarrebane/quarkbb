package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"codeberg.org/ronia/quarkbb/internal/model"
	"codeberg.org/ronia/quarkbb/internal/repository/postgres"
)

type sqlUserRepository struct {
	q *postgres.Queries
}

// NewUserRepository creates a new UserRepository implementation using SQLC queries.
func NewUserRepository(q *postgres.Queries) UserRepository {
	return &sqlUserRepository{
		q: q,
	}
}

// CreateUser persists a new user to the database and returns the created user with ID.
func (ur sqlUserRepository) CreateUser(ctx context.Context, user model.RegisterUserCommand) (*model.User, error) {
	cmd := postgres.CreateUserAndReturnIdParams{
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
func (ur sqlUserRepository) UsernameExists(ctx context.Context, username string) (bool, error) {
	count, err := ur.q.CountUserByUsername(ctx, username)
	if err != nil {
		return false, fmt.Errorf("count user by username: %w", err)
	}

	return count > 0, nil
}

// GetUserByID retrieves a user from the database by their id.
// Returns the user model if found. If not, NotFoundError
func (ur sqlUserRepository) GetUserByID(ctx context.Context, id int64) (*model.User, error) {
	user, err := ur.q.GetUserById(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &NotFoundError{Entity: "user", Key: fmt.Sprintf("%d", id)}
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	return &model.User{
		ID:       user.ID,
		Username: user.Username,
		Password: user.Password,
		Email:    user.Email,
	}, nil
}

// FindUserByUsername retrieves a user from the database by their username.
// Returns the user model if found.
func (ur sqlUserRepository) FindUserByUsername(ctx context.Context, username string) (*model.User, error) {
	user, err := ur.q.FindUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &NotFoundError{Entity: "user", Key: username}
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	return &model.User{
		ID:       user.ID,
		Username: user.Username,
		Password: user.Password,
		Email:    user.Email,
	}, nil
}
