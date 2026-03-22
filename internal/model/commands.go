// Package model contains all domain models and data transfer objects.
// It defines the core data structures used throughout the application,
// including user entities, command objects, and request/response models.
package model

// RegisterUserCommand represents the data required for user registration.
type RegisterUserCommand struct {
	Username string `json:"username" validate:"required,printascii,min=3"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=5"`
}

// LoginCommand represents the data required for login.
type LoginCommand struct {
	Username string `json:"username" validate:"required,printascii,min=3"`
	Password string `json:"password" validate:"required,min=5"`
}
