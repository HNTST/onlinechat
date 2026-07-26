package chats_postgres_repository

import (
	"context"
	"discord/internal/db"
	"fmt"
)

func (r *ChatRepository) UpdateChatName(
	ctx context.Context,
	chatID int64,
	newName string,
) error {
	q := r.pool.GetQueries(ctx)
	err := q.UpdateChatNameByID(ctx, db.UpdateChatNameByIDParams{
		Name: newName,
		ID:   chatID,
	})
	if err != nil {
		return fmt.Errorf("update chat name: %w", err)
	}
	return nil
}
