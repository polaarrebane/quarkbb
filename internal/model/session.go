package model

import "time"

// SessionStatus represents the status of the session
type SessionStatus string

// Session statuses
const (
	SessionActive SessionStatus = "active"
	SessionClosed SessionStatus = "closed"
)

// Session represents a session entity in the system.
type Session struct {
	ID        int64
	UserID    int64
	PublicID  string
	Status    SessionStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}
