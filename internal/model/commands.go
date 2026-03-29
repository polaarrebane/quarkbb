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

// RefreshCommand represents the data required for refresh an auth token.
type RefreshCommand struct {
	UserID    int64  `json:"userid"`
	SessionID string `json:"sessionid"`
	JTI       string `json:"jti"`
}

// LogoutCommand represents the data required for logout.
type LogoutCommand struct {
	UserID    int64  `json:"userid"`
	SessionID string `json:"sessionid"`
	JTI       string `json:"jti"`
}

// RetrieveSessionsCommand represents the data required for retrieving all sessions.
type RetrieveSessionsCommand struct {
	Username string `json:"userid"`
}
