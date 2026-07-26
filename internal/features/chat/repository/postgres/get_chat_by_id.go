package chats_postgres_repository

import (
	"context"
	"discord/internal/core/domain"
	core_errors "discord/internal/core/errors"
	"discord/internal/db"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (r *ChatRepository) GetChatByID(ctx context.Context, userID, chatID int64) (domain.Chat, error) {
	q := r.pool.GetQueries(ctx)
	c, err := q.GetChatWithMembership(ctx, db.GetChatWithMembershipParams{
		ID:     chatID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Chat{}, fmt.Errorf("chat not found or user not a member: %w", core_errors.ErrChatNotExists)
		}
		return domain.Chat{}, fmt.Errorf("get chat by id: %w", err)
	}
	return domain.Chat{Id: c.ID, Name: c.Name, IsGroup: c.IsGroup}, nil
}
