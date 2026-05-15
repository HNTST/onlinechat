package chats_postgres_repository

import (
	"context"
	core_postgres_pool "discord/internal/core/repository/postgres/pool"
)

type ChatRepository struct {
	pool core_postgres_pool.Pool
}

func NewChatRepository(
	pool core_postgres_pool.Pool,
) *ChatRepository {
	return &ChatRepository{
		pool: pool,
	}
}

func (r *ChatRepository) Transaction(
	txCtx context.Context,
	fn func(ctx context.Context) error,
) error {
	return r.pool.CtxWithTransaction(txCtx, fn)
}
