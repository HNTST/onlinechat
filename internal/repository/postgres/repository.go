package postgres

import core_postgres_pool "discord/internal/core/repository/postgres/pool"

type OldDbREpository struct {
	pool core_postgres_pool.Pool
}

func NewOldRepostiroty(pool core_postgres_pool.Pool) *OldDbREpository {
	return &OldDbREpository{
		pool: pool,
	}
}
