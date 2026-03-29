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

// MalformedKeyError is returned when a requested key is malformed (i.e. not an uuid).
type MalformedKeyError struct {
	Key string
}

func (e *MalformedKeyError) Error() string {
	return fmt.Sprintf("malformed key: %s", e.Key)
}
