package chats_postgres_repository

import (
	"context"
	"fmt"
)

func (r *ChatRepository) GetUserChatIDs(ctx context.Context, userID int64) ([]int64, error) {
	q := r.pool.GetQueries(ctx)
	result, err := q.GetUserChatIDs(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	return result, nil
}
