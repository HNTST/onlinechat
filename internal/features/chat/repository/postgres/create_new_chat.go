package chats_postgres_repository

import (
	"context"
	"discord/internal/db"
	"fmt"
)

func (r *ChatRepository) CreateNewChat(
	ctx context.Context,
	chatName string,
	isGroup bool,
) (int64, error) {
	q := r.pool.GetQueries(ctx)
	chatID, err := q.InsertChat(ctx, db.InsertChatParams{
		Name:    chatName,
		IsGroup: isGroup,
	})
	if err != nil {
		return 0, fmt.Errorf("scan error: %w", err)
	}
	return chatID, nil
}
