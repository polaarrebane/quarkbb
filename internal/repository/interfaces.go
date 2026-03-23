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
	GetUserByID(ctx context.Context, id int64) (*model.User, error)
	FindUserByUsername(ctx context.Context, username string) (*model.User, error)
}

// RefreshTokenRepository defines the interface for refresh token data access operations.
type RefreshTokenRepository interface {
	RefreshTokenUsed(ctx context.Context, jti string) (bool, error)
	MarkTokenAsUsed(ctx context.Context, jti string) error
}
