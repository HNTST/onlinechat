package chats_postgres_repository

import (
	"context"
	core_errors "discord/internal/core/errors"
	"discord/internal/db"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

func (r *ChatRepository) SoftDeleteMessage(ctx context.Context, chatID, messageID int64) error {
	query := `
		UPDATE messages
		SET is_deleted = true, message = ''
		WHERE id = $1 AND chat_id = $2 AND is_deleted = false
	`

	var (
		tag pgconn.CommandTag
		err error
	)

	tx := r.pool.GetTx(ctx)
	if tx != nil {
		tag, err = tx.Exec(ctx, query, messageID, chatID)
	} else {
		ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
		defer cancel()
		tag, err = r.pool.Exec(ctx, query, messageID, chatID)
	}

	if err != nil {
		return fmt.Errorf("exec error: %w", err)
	}

	if tag.RowsAffected() == 0 {
		q := r.pool.GetQueries(ctx)

		inChat, checkErr := q.CheckMessageExistsInChat(ctx, db.CheckMessageExistsInChatParams{
			ID:     messageID,
			ChatID: chatID,
		})
		if checkErr != nil {
			return fmt.Errorf("message not found: %w", core_errors.ErrMessageNotFound)
		}
		if inChat {
			return fmt.Errorf("message already deleted: %w", core_errors.ErrMessageNotFound)
		}

		exists, existErr := q.CheckMessageExists(ctx, messageID)
		if existErr != nil || !exists {
			return fmt.Errorf("message not found: %w", core_errors.ErrMessageNotFound)
		}
		return fmt.Errorf("message belongs to another chat: %w", core_errors.ErrUnauthorized)
	}
	return nil
}
