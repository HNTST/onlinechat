package chats_postgres_repository

import (
	"context"
	"discord/internal/core/domain"
	"fmt"
)

func (r *ChatRepository) GetChatMessages(ctx context.Context, chat_id int64) ([]domain.Messages, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		select sender_id, message, sent_at
		from messages
		where chat_id = $1
		ORDER BY id ASC
	`
	rows, err := r.pool.Query(ctx, query, chat_id)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	var result []domain.Messages
	for rows.Next() {
		var m domain.Messages
		if err := rows.Scan(&m.SenderId, &m.Message, &m.Sent_at); err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		result = append(result, m)
	}
	return result, nil
}
