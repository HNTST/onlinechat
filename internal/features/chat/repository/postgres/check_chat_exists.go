package chats_postgres_repository

import (
	"context"
	core_errors "discord/internal/core/errors"
	"discord/internal/db"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (r *ChatRepository) CheckExistsChat(
	ctx context.Context,
	userID, targetUserID int64,
) (int64, error) {
	q := r.pool.GetQueries(ctx)
	chatID, err := q.CheckChatExistsBetweenUsers(ctx, db.CheckChatExistsBetweenUsersParams{
		UserID:   userID,
		UserID_2: targetUserID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, fmt.Errorf("get rows: no rows: %w", core_errors.ErrChatNotExists)
		}
		return 0, fmt.Errorf("get rows: %w", err)
	}
	return chatID, nil
}
