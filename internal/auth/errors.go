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
)
