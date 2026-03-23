package repository

import "fmt"

// NotFoundError is returned when a requested entity does not exist.
type NotFoundError struct {
	Entity string // e.g. "user"
	Key    string // e.g. the username that was looked up
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found: %s", e.Entity, e.Key)
}
