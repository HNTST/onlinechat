package chats_postgres_repository

import (
	"context"
	core_errors "discord/internal/core/errors"
	"fmt"
)

func (r *ChatRepository) SoftDeleteMessage(ctx context.Context, messageID int64) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		UPDATE messages
		SET is_deleted = true, message = ''
		WHERE id = $1 AND is_deleted = false
	`
	tag, err := r.pool.Exec(ctx, query, messageID)
	if err != nil {
		return fmt.Errorf("exec error: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("message not found or already deleted: %w", core_errors.ErrMessageNotFound)
	}
	return nil
}
