package chats_postgres_repository

import (
	"context"
	"discord/internal/core/domain"
	"fmt"
)

func (r *ChatRepository) GetUserChats(ctx context.Context, userID int64) ([]domain.Chat, error) {
	q := r.pool.GetQueries(ctx)
	rows, err := q.GetChatsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	result := make([]domain.Chat, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.Chat{Id: row.ID, Name: row.Name, IsGroup: row.IsGroup})
	}
	return result, nil
}
