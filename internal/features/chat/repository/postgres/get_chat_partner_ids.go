package chats_postgres_repository

import (
	"context"
	"fmt"
)

func (r *ChatRepository) GetChatPartnerIDs(ctx context.Context, userID int64) ([]int64, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT DISTINCT cm2.user_id
		FROM chat_members cm1
		JOIN chat_members cm2 ON cm1.chat_id = cm2.chat_id
		WHERE cm1.user_id = $1 AND cm2.user_id != $1
	`
	rows, err := r.pool.Query(ctx, query, userID)
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
