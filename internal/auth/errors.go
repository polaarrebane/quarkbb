package auth

import (
	"errors"
)

var (
	errMalformedBody      = errors.New("malformed request body")
	errValidationError    = errors.New("validation error")
	errInternalError      = errors.New("internal error")
	errRegistrationFailed = errors.New("registration failed")
	errUserAlreadyExists  = errors.New("user already exists")
	errUserNotFound       = errors.New("user not found")
	errWrongPassword      = errors.New("wrong password")
)
