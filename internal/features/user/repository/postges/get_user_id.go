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

	query := `select id from users where username=$1`
	row := r.pool.QueryRow(ctx, query, username)
	var id int64
	err := row.Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, fmt.Errorf("user not found: %w", core_errors.ErrUserNotFound)
		}
		return 0, fmt.Errorf("scan error: %w", err)
	}
	return id, nil
}
