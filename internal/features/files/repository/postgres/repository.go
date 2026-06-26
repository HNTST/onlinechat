package files_postgres_repository

import core_postgres_pool "discord/internal/core/repository/postgres/pool"

type FileRepository struct {
	pool core_postgres_pool.Pool
}

func NewFileRepository(pool core_postgres_pool.Pool) *FileRepository {
	return &FileRepository{
		pool: pool,
	}
}
