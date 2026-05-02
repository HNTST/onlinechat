package users_postgres_repository

import core_postgres_pool "discord/internal/core/repository/postgres/pool"

type UserRepository struct {
	pool core_postgres_pool.Pool
}

func NewUsersRepository(pool core_postgres_pool.Pool) *UserRepository {
	return &UserRepository{
		pool: pool,
	}
}
