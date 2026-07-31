package users_postgres_repository

import (
	"context"
	"discord/internal/db"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

func (r *UserRepository) UpdateAvatar(ctx context.Context, userID, fileID int64) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	q := r.pool.GetQueries(ctx)
	err := q.UpdateUserAvatarByID(ctx, db.UpdateUserAvatarByIDParams{
		AvatarID: pgtype.Int8{Int64: fileID, Valid: true},
		ID:       userID,
	})
	if err != nil {
		return fmt.Errorf("update user avatar: %w", err)
	}
	return nil
}
