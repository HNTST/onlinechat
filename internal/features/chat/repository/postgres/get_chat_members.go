package chats_postgres_repository

import (
	"context"
	"fmt"
)

func (r *ChatRepository) GetChatMembers(ctx context.Context, chat_id int64) ([]int64, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		select user_id
		from chat_members
		where chat_id = $1
	`
	rows, err := r.pool.Query(ctx, query, chat_id)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	var result []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		result = append(result, id)
	}
	return result, nil
}
