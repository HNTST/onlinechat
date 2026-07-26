package chats_postgres_repository

import (
	"context"
	core_errors "discord/internal/core/errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

func (r *ChatRepository) EditMessage(ctx context.Context, messageID int64, text string) error {
	query := `
		UPDATE messages
		SET message = $1, is_edited = true
		WHERE id = $2 AND is_deleted = false
	`

	tx := r.pool.GetTx(ctx)

	var (
		tag pgconn.CommandTag
		err error
	)

	if tx != nil {
		tag, err = tx.Exec(ctx, query, text, messageID)
	} else {
		ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
		defer cancel()
		tag, err = r.pool.Exec(ctx, query, text, messageID)
	}

	if err != nil {
		return fmt.Errorf("exec error: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("message not found: %w", core_errors.ErrMessageNotFound)
	}
	return nil
}
