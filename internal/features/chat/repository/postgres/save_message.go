package chats_postgres_repository

import (
	"context"
	"discord/internal/core/domain"
	"fmt"
)

func (r *ChatRepository) SaveMessage(ctx context.Context, message domain.Messages) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		insert into messages (chat_id, sender_id, message, sent_at)
		values($1, $2, $3, $4)
	`
	_, err := r.pool.Exec(ctx, query, message.ChatId, message.SenderId, message.Message, message.Sent_at)
	if err != nil {
		return fmt.Errorf("exec error: %w", err)
	}
	return nil
}
