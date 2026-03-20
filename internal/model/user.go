// Package model contains all domain models and commands
package model

// User represents a user entity in the system.
// This model is used for storing user data in the database and
// for transferring user information between application layers.
// Note: The Password field contains a bcrypt hash, not plain text.
type User struct {
	ID       int64
	Username string
	Email    string
	Password string
}
