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

func int64SliceToInt32(s []int64) []int32 {
	result := make([]int32, len(s))
	for i, v := range s {
		result[i] = int32(v)
	}
	return result
}
