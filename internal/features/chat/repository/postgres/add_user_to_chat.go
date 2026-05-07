package chats_postgres_repository

import (
	"context"
	"fmt"
)

func (r *ChatRepository) AddUserToChat(
	ctx context.Context,
	chat_id, user_id int64,
) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		insert into chat_members (chat_id, user_id)
		values($1, $2)
	`
	_, err := r.pool.Exec(ctx, query, chat_id, user_id)
	if err != nil {
		return fmt.Errorf("add user to chat: %w", err)
	}
	return nil
}
