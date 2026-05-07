package chats_postgres_repository

import core_postgres_pool "discord/internal/core/repository/postgres/pool"

type ChatRepository struct {
	pool core_postgres_pool.Pool
}

func NewChatRepository(pool core_postgres_pool.Pool) *ChatRepository {
	return &ChatRepository{
		pool: pool,
	}
}
