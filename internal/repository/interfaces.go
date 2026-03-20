// Package repository defines the data access interfaces and implementations
package repository

import (
	"context"

	"codeberg.org/ronia/quarkbb/internal/model"
)

// UserRepository is an interface of repository of users
type UserRepository interface {
	CreateUser(ctx context.Context, user model.RegisterUserCommand) (*model.User, error)
	UsernameExists(ctx context.Context, username string) (bool, error)
}
