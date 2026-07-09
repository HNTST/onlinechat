package chats_postgres_repository

import (
	"context"
	"fmt"
)

func (r *ChatRepository) UpdateChatName(
	ctx context.Context,
	chatID int64,
	newName string,
) error {
	var err error
	query := `
		update chats
		set name = $1
		where id = $2
	`

	tx := r.pool.GetTx(ctx)

	if tx != nil {
		_, err = tx.Exec(ctx, query, newName, chatID)
	} else {
		ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
		defer cancel()
		_, err = r.pool.Exec(ctx, query, newName, chatID)
	}

	if err != nil {
		return fmt.Errorf("update chat name: %w", err)
	}
	return nil
}
