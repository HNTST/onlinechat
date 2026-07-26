package chats_postgres_repository

import (
	"context"
	"fmt"
)

func (r *ChatRepository) DeleteChat(
	ctx context.Context,
	chatID int64,
) error {
	q := r.pool.GetQueries(ctx)
	err := q.DeleteChatByID(ctx, chatID)
	if err != nil {
		return fmt.Errorf("delete chat: %w", err)
	}
	return nil
}
