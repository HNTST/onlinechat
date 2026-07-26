package auth_postgres_repository

import (
	"context"
	"discord/internal/core/domain"
	core_errors "discord/internal/core/errors"
	"discord/internal/db"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

func (r *AuthRepository) CreateSession(ctx context.Context, userID int64, token string, expiresAt time.Time) (domain.Session, error) {
	q := r.pool.GetQueries(ctx)
	s, err := q.CreateSession(ctx, db.CreateSessionParams{
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return domain.Session{}, fmt.Errorf("create session: %w", err)
	}
	return domain.Session{
		ID:        s.ID,
		UserID:    s.UserID,
		Token:     s.Token,
		ExpiresAt: s.ExpiresAt,
	}, nil
}

func (r *AuthRepository) GetSessionByToken(ctx context.Context, token string) (domain.Session, error) {
	q := r.pool.GetQueries(ctx)
	s, err := q.GetSessionByToken(ctx, token)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Session{}, fmt.Errorf("session not found: %w", core_errors.ErrUnauthorized)
		}
		return domain.Session{}, fmt.Errorf("get session: %w", err)
	}
	return domain.Session{
		ID:        s.ID,
		UserID:    s.UserID,
		Token:     s.Token,
		ExpiresAt: s.ExpiresAt,
	}, nil
}

func (r *AuthRepository) DeleteSession(ctx context.Context, token string) error {
	q := r.pool.GetQueries(ctx)
	err := q.DeleteSession(ctx, token)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}
