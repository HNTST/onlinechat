package files_postgres_repository

import (
	"context"
	core_postgres_pool "discord/internal/core/repository/postgres/pool"
)

type FileRepository struct {
	pool core_postgres_pool.Pool
}

func NewFileRepository(pool core_postgres_pool.Pool) *FileRepository {
	return &FileRepository{
		pool: pool,
	}
}

func (r *FileRepository) Transaction(ctx context.Context, fn func(txCtx context.Context) error) error {
	return r.pool.CtxWithTransaction(ctx, fn)
}
