package users_postgres_repository

import (
	"context"
	"discord/internal/core/domain"
	"discord/internal/db"
	"fmt"
)

func (r *UserRepository) SaveUser(ctx context.Context, user domain.User) (domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	q := r.pool.GetQueries(ctx)
	row, err := q.InsertUser(ctx, db.InsertUserParams{
		Username: user.Name,
		Password: user.Password,
		Reg:      user.Reg,
	})
	if err != nil {
		return domain.User{}, fmt.Errorf("query error: %w", err)
	}

	var avatarID *int64
	if row.AvatarID.Valid {
		v := row.AvatarID.Int64
		avatarID = &v
	}

	return domain.NewUserWithAvatar(row.ID, row.Username, row.Password, row.Reg, avatarID), nil
}
