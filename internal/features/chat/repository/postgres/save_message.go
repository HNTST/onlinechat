package chats_postgres_repository

import (
	"context"
	"discord/internal/core/domain"
	"discord/internal/db"
	"fmt"
)

func (r *ChatRepository) SaveMessage(ctx context.Context, message domain.Messages) (int64, error) {
	q := r.pool.GetQueries(ctx)
	id, err := q.InsertMessage(ctx, db.InsertMessageParams{
		ChatID:   message.ChatId,
		SenderID: message.SenderId,
		Message:  message.Message,
		SentAt:   message.Sent_at,
	})
	if err != nil {
		return 0, fmt.Errorf("exec error: %w", err)
	}
	return id, nil
}
