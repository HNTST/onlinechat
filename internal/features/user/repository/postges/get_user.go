package users_postgres_repository

import (
	"context"
	"discord/internal/core/domain"
	"fmt"
)

func (r *UserRepository) GetUser(ctx context.Context, user_id int64) (domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		select id, username, reg
		from users
		where id=$1
	`
	row := r.pool.QueryRow(ctx, query, user_id)
	var userModel UserModel
	err := row.Scan(
		&userModel.ID,
		&userModel.Username,
		&userModel.Reg,
	)
	if err != nil {
		return domain.User{}, fmt.Errorf("scan error: %w", err)
	}
	userDomain := domain.NewUser(
		userModel.ID,
		"-1",
		userModel.Username,
		userModel.Reg,
	)
	return userDomain, err

}
