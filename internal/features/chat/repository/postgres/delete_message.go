package chats_postgres_repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

func (r *ChatRepository) SoftDeleteMessage(ctx context.Context, chatID, messageID, userID int64) error {
	query := `
		UPDATE messages
		SET is_deleted = true, message = ''
		WHERE id = $1 AND chat_id = $2 AND sender_id = $3 AND is_deleted = false
	`

	var (
		tag pgconn.CommandTag
		err error
	)

	tx := r.pool.GetTx(ctx)
	if tx != nil {
		tag, err = tx.Exec(ctx, query, messageID, chatID, userID)
	} else {
		ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
		defer cancel()
		tag, err = r.pool.Exec(ctx, query, messageID, chatID, userID)
	}

	if err != nil {
		return fmt.Errorf("exec error: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("soft delete message: %w", r.checkMessageAccess(ctx, messageID, chatID, userID))
	}
	return nil
}
