package chats_postgres_repository

import (
	"context"
	"fmt"
)

func (r *ChatRepository) DeleteChat(
	ctx context.Context,
	chatID int64,
) error {
	var err error
	query := `
		delete from chats
		where id = $1
	`

	tx := r.pool.GetTx(ctx)

	if tx != nil {
		_, err = tx.Exec(ctx, query, chatID)
	} else {
		ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
		defer cancel()
		_, err = r.pool.Exec(ctx, query, chatID)
	}

	if err != nil {
		return fmt.Errorf("delete chat: %w", err)
	}
	return nil
}
