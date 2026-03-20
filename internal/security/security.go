// Package security provides cryptographic and security-related utilities.
// It includes functions for password hashing using bcrypt and other
// security operations needed for user authentication and data protection.
package security

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

// GeneratePasswordHash creates a secure bcrypt hash from a plain text password.
func GeneratePasswordHash(p string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(p), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("generate password hash: %w", err)
	}
	return string(hash), nil
}
