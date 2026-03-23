package repository

import (
	"context"
	"fmt"

	"codeberg.org/ronia/quarkbb/internal/repository/postgres"
	"github.com/google/uuid"
)

type sqlRefreshTokenRepository struct {
	q *postgres.Queries
}

// NewRefreshTokenRepository creates a new RefreshTokenRepository implementation using SQLC queries.
func NewRefreshTokenRepository(q *postgres.Queries) RefreshTokenRepository {
	return &sqlRefreshTokenRepository{
		q: q,
	}
}

func (r *sqlRefreshTokenRepository) RefreshTokenUsed(ctx context.Context, jti string) (bool, error) {
	id, err := uuid.Parse(jti)
	if err != nil {
		return false, fmt.Errorf("parse jti: %w", err)
	}
	count, err := r.q.CountUsedRefreshTokenByJTI(ctx, id)
	if err != nil {
		return false, fmt.Errorf("count refresh_tokens by jti: %w", err)
	}
	return count > 0, nil
}

func (r *sqlRefreshTokenRepository) MarkTokenAsUsed(ctx context.Context, jti string) error {
	id, err := uuid.Parse(jti)
	if err != nil {
		return fmt.Errorf("parse jti: %w", err)
	}

	if err := r.q.MarkTokenAsUsed(ctx, id); err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	return nil
}
