package auth_postgres_repository

import (
	"context"
	"discord/internal/core/domain"
	core_errors "discord/internal/core/errors"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

func (r *AuthRepository) CreateSession(ctx context.Context, userID int64, token string, expiresAt time.Time) (domain.Session, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		insert into sessions (user_id, token, expires_at)
		values($1,$2,$3)
		returning id, user_id, token, expires_at
	`
	row := r.pool.QueryRow(ctx, query, userID, token, expiresAt)
	var model SessionModel
	err := row.Scan(
		&model.ID,
		&model.UserID,
		&model.Token,
		&model.ExpiresAt,
	)
	if err != nil {
		return domain.Session{}, fmt.Errorf("create session: %w", err)
	}
	return domain.Session{
		ID:        model.ID,
		UserID:    model.UserID,
		Token:     model.Token,
		ExpiresAt: model.ExpiresAt,
	}, nil
}

func (r *AuthRepository) GetSessionByToken(ctx context.Context, token string) (domain.Session, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		select id, user_id, token, expires_at
		from sessions
		where token=$1
	`
	row := r.pool.QueryRow(ctx, query, token)
	var model SessionModel
	err := row.Scan(
		&model.ID,
		&model.UserID,
		&model.Token,
		&model.ExpiresAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Session{}, fmt.Errorf("session not found: %w", core_errors.ErrUnauthorized)
		}
		return domain.Session{}, fmt.Errorf("get session: %w", err)
	}
	return domain.Session{
		ID:        model.ID,
		UserID:    model.UserID,
		Token:     model.Token,
		ExpiresAt: model.ExpiresAt,
	}, nil
}

func (r *AuthRepository) DeleteSession(ctx context.Context, token string) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `delete from sessions where token=$1`
	_, err := r.pool.Exec(ctx, query, token)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}
