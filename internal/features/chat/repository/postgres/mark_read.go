package chats_postgres_repository

import (
	"context"
	"fmt"
)

func (r *ChatRepository) MarkMessagesAsRead(
	ctx context.Context,
	chatID int64,
	userID int64,
	lastReadMessageID int64,
) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		UPDATE messages
		SET is_read = true
		WHERE chat_id = $1 AND sender_id != $2 AND id <= $3 AND is_read = false
	`
	_, err := r.pool.Exec(ctx, query, chatID, userID, lastReadMessageID)
	if err != nil {
		return fmt.Errorf("exec error: %w", err)
	}
	return nil
}
