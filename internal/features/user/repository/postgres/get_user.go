package users_postgres_repository

import (
	"context"
	"discord/internal/core/domain"
	core_errors "discord/internal/core/errors"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (r *UserRepository) GetUser(ctx context.Context, user_id int64) (domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		select id, username, reg, avatar_id
		from users
		where id=$1
	`
	row := r.pool.QueryRow(ctx, query, user_id)
	var userModel UserModel
	err := row.Scan(
		&userModel.ID,
		&userModel.Username,
		&userModel.Reg,
		&userModel.AvatarID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, fmt.Errorf("user not found: %w", core_errors.ErrUserNotFound)
		}
		return domain.User{}, fmt.Errorf("scan error: %w", err)
	}
	userDomain := domain.NewUserWithAvatar(
		userModel.ID,
		userModel.Username,
		"-1",
		userModel.Reg,
		userModel.AvatarID,
	)
	return userDomain, err

}
