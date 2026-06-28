package chats_postgres_repository

import (
	"context"
	core_errors "discord/internal/core/errors"
	"fmt"
)

func (r *ChatRepository) EditMessage(ctx context.Context, messageID int64, text string) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		UPDATE messages
		SET message = $1, is_edited = true
		WHERE id = $2 AND is_deleted = false
	`
	tag, err := r.pool.Exec(ctx, query, text, messageID)
	if err != nil {
		return fmt.Errorf("exec error: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("message not found: %w", core_errors.ErrMessageNotFound)
	}
	return nil
}
