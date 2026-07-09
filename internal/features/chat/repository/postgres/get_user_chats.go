package chats_postgres_repository

import (
	"context"
	"discord/internal/core/domain"
	"fmt"
)

func (r *ChatRepository) GetUserChats(ctx context.Context, user_id int64) ([]domain.Chat, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT c.id, c.name, c.is_group
		FROM chats c
		JOIN chat_members cm ON c.id = cm.chat_id
		WHERE cm.user_id = $1
	`
	rows, err := r.pool.Query(ctx, query, user_id)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	var result []domain.Chat
	for rows.Next() {
		var c domain.Chat
		if err := rows.Scan(&c.Id, &c.Name, &c.IsGroup); err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		result = append(result, c)
	}
	return result, nil
}
