package users_postgres_repository

import (
	"context"
	"discord/internal/core/domain"
	core_errors "discord/internal/core/errors"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (r *UserRepository) GetUser(ctx context.Context, userID int64) (domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	q := r.pool.GetQueries(ctx)
	row, err := q.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, fmt.Errorf("user not found: %w", core_errors.ErrUserNotFound)
		}
		return domain.User{}, fmt.Errorf("query error: %w", err)
	}

	var avatarID *int64
	if row.AvatarID.Valid {
		v := row.AvatarID.Int64
		avatarID = &v
	}

	return domain.NewUserWithAvatar(row.ID, row.Username, row.Password, row.Reg, avatarID), nil
}
