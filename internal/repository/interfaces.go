// Package repository defines the data access interfaces and implementations
package repository

import (
	"context"

	"codeberg.org/ronia/quarkbb/internal/model"
)

// UserRepository defines the interface for user data access operations.
type UserRepository interface {
	CreateUser(ctx context.Context, user model.RegisterUserCommand) (*model.User, error)
	UsernameExists(ctx context.Context, username string) (bool, error)
	FindUserByUsername(ctx context.Context, username string) (*model.User, error)
}
