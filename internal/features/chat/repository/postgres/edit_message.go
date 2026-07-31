package chats_postgres_repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

func (r *ChatRepository) EditMessage(ctx context.Context, chatID, messageID, userID int64, text string) error {
	query := `
		UPDATE messages
		SET message = $1, is_edited = true
		WHERE id = $2 AND chat_id = $3 AND sender_id = $4 AND is_deleted = false
	`

	tx := r.pool.GetTx(ctx)

	var (
		tag pgconn.CommandTag
		err error
	)

	if tx != nil {
		tag, err = tx.Exec(ctx, query, text, messageID, chatID, userID)
	} else {
		ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
		defer cancel()
		tag, err = r.pool.Exec(ctx, query, text, messageID, chatID, userID)
	}

	if err != nil {
		return fmt.Errorf("exec error: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("edit message: %w", r.checkMessageAccess(ctx, messageID, chatID, userID))
	}
	return nil
}
