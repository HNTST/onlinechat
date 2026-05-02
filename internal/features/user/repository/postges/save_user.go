package users_postgres_repository

import (
	"context"
	"discord/internal/core/domain"
	"fmt"
)

func (r *UserRepository) SaveUser(ctx context.Context, user domain.User) (domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
			insert into users (username, password,reg)
			values($1,$2,$3)
			RETURNING id, username, reg
		`

	row := r.pool.QueryRow(ctx, query, user.Name, user.Password)
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
