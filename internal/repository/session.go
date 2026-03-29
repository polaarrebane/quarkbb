package repository

import (
	"context"
	"fmt"
	"time"

	"codeberg.org/ronia/quarkbb/internal/model"
	"codeberg.org/ronia/quarkbb/internal/repository/postgres"
	"github.com/google/uuid"
)

type sessionRepository struct {
	q *postgres.Queries
}

// NewSessionRepository creates a new SessionRepository implementation using SQLC queries.
func NewSessionRepository(q *postgres.Queries) SessionRepository {
	return &sessionRepository{
		q: q,
	}
}

func (r *sessionRepository) CreateSession(ctx context.Context, user *model.User) (*model.Session, error) {
	now := time.Now()
	publicID := uuid.New()
	params := postgres.CreateSessionParams{
		UserID:    user.ID,
		PublicID:  publicID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	sessionID, err := r.q.CreateSession(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	result := &model.Session{
		ID:        sessionID,
		UserID:    user.ID,
		PublicID:  publicID.String(),
		Status:    model.SessionActive,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return result, nil
}

func (r *sessionRepository) GetAllSessions(ctx context.Context, user *model.User) ([]model.Session, error) {
	rows, err := r.q.GetAllSessionsByUserId(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	items := make([]model.Session, len(rows))
	for i, row := range rows {
		items[i] = model.Session{
			ID:        row.ID,
			UserID:    row.UserID,
			PublicID:  row.PublicID.String(),
			Status:    model.SessionStatus(row.Status),
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		}
	}
	return items, nil
}

func (r *sessionRepository) IsSessionClosed(ctx context.Context, publicID string) (bool, error) {
	id, err := uuid.Parse(publicID)
	if err != nil {
		return false, fmt.Errorf("parse public id of session: %w", err)
	}
	count, err := r.q.CountClosedSessionByPublicID(ctx, id)
	if err != nil {
		return false, fmt.Errorf("count closed sessions by public id: %w", err)
	}
	return count > 0, nil
}

func (r *sessionRepository) CloseSession(ctx context.Context, publicID string) error {
	id, err := uuid.Parse(publicID)
	if err != nil {
		return fmt.Errorf("parse public id of session: %w", err)
	}
	if err := r.q.CloseSession(ctx, id); err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	return nil
}
