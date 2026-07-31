package chats_postgres_repository

import (
	"context"
	core_errors "discord/internal/core/errors"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (r *ChatRepository) checkMessageAccess(ctx context.Context, messageID, chatID, userID int64) error {
	q := r.pool.GetQueries(ctx)
	owner, err := q.GetMessageOwner(ctx, messageID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("message not found: %w", core_errors.ErrMessageNotFound)
		}
		return fmt.Errorf("get message owner: %w", err)
	}

	if owner.ChatID != chatID {
		return fmt.Errorf("message belongs to another chat: %w", core_errors.ErrUnauthorized)
	}
	if owner.IsDeleted {
		return fmt.Errorf("message already deleted: %w", core_errors.ErrMessageNotFound)
	}
	if owner.SenderID != userID {
		return fmt.Errorf("message not owned by user: %w", core_errors.ErrUnauthorized)
	}
	return fmt.Errorf("message access denied: %w", core_errors.ErrMessageNotFound)
}
