package auth_postgres_repository

import (
	"context"
	core_postgres_pool "discord/internal/core/repository/postgres/pool"
)

type AuthRepository struct {
	pool core_postgres_pool.Pool
}

func NewAuthRepository(pool core_postgres_pool.Pool) *AuthRepository {
	return &AuthRepository{
		pool: pool,
	}
}

func (r *AuthRepository) Transaction(txCtx context.Context, fn func(ctx context.Context) error) error {
	return r.pool.CtxWithTransaction(txCtx, fn)
}
