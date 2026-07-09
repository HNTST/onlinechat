package chats_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (r *ChatRepository) IsUserInChat(
	ctx context.Context,
	userID int64,
	chatID int64,
) (bool, error) {
	var (
		err    error
		exists bool
		row    pgx.Row
	)
	query := `
		select exists(
			select 1
			from chat_members
			where chat_id = $1 and user_id = $2
		)
	`

	tx := r.pool.GetTx(ctx)

	if tx != nil {
		row = tx.QueryRow(ctx, query, chatID, userID)
	} else {
		ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
		defer cancel()
		row = r.pool.QueryRow(ctx, query, chatID, userID)
	}

	err = row.Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("is user in chat: %w", err)
	}
	return exists, nil
}
