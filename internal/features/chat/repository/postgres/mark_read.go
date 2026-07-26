package chats_postgres_repository

import (
	"context"
	"discord/internal/db"
	"fmt"
)

func (r *ChatRepository) MarkMessagesAsRead(
	ctx context.Context,
	chatID int64,
	userID int64,
	lastReadMessageID int64,
) error {
	q := r.pool.GetQueries(ctx)
	err := q.MarkMessagesReadByChatID(ctx, db.MarkMessagesReadByChatIDParams{
		ChatID:   chatID,
		SenderID: userID,
		ID:       lastReadMessageID,
	})
	if err != nil {
		return fmt.Errorf("exec error: %w", err)
	}
	return nil
}
