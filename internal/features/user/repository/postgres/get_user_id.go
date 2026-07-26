package users_postgres_repository

import (
	"context"
	core_errors "discord/internal/core/errors"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (r *UserRepository) GetUserID(ctx context.Context, username string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	q := r.pool.GetQueries(ctx)
	id, err := q.GetUserIDByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, fmt.Errorf("user not found: %w", core_errors.ErrUserNotFound)
		}
		return 0, fmt.Errorf("query error: %w", err)
	}
	return id, nil
}
