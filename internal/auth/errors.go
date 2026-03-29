package auth

// authError is an unexported type so sentinel values cannot be
// created or compared outside this package.
type authError string

func (e authError) Error() string { return string(e) }

const (
	errMalformedBody           authError = "malformed request body"
	errValidationError         authError = "validation error"
	errInternalError           authError = "internal error"
	errRegistrationFailed      authError = "registration failed"
	errUserAlreadyExists       authError = "user already exists"
	errLoginFailed             authError = "login failed"
	errUserNotFound            authError = "user not found"
	errWrongPassword           authError = "wrong password"
	errRefreshFailed           authError = "refresh failed"
	errAuthorizationRequired   authError = "authorization required"
	errSessionNotFound         authError = "session not found"
	errMalformedID             authError = "malformed id"
	errCantCloseCurrentSession authError = "current session can't be closed"
)
