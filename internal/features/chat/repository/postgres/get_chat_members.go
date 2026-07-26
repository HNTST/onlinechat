package chats_postgres_repository

import (
	"context"
	"fmt"
)

func (r *ChatRepository) GetChatMembers(ctx context.Context, chatID int64) ([]int64, error) {
	q := r.pool.GetQueries(ctx)
	result, err := q.GetChatMembersByChatID(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	return result, nil
}
