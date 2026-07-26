package chats_postgres_repository

import (
	"context"
	"discord/internal/db"
	"fmt"
)

func (r *ChatRepository) AddUserToChat(
	ctx context.Context,
	chatID int64,
	userIDs []int64,
) error {
	q := r.pool.GetQueries(ctx)
	err := q.InsertChatMember(ctx, db.InsertChatMemberParams{
		ChatID:  chatID,
		Column2: int64SliceToInt32(userIDs),
	})
	if err != nil {
		return fmt.Errorf("add user to chat: %w", err)
	}
	return nil
}
