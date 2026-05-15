package chats_postgres_repository

import (
	"context"
	"fmt"
)

func (r *ChatRepository) AddUserToChat(
	ctx context.Context,
	chat_id int64,
	users_ids []int64,
) error {
	var err error
	query := `
		insert into chat_members (chat_id, user_id)
		select $1, unnest($2::integer[])
		on conflict (chat_id, user_id) do nothing;
	`

	tx := r.pool.GetTx(ctx)

	if tx != nil {
		_, err = tx.Exec(ctx, query, chat_id, users_ids)
	} else {
		ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
		defer cancel()
		_, err = r.pool.Exec(ctx, query, chat_id, users_ids)
	}

	if err != nil {
		return fmt.Errorf("add user to chat: %w", err)
	}
	return nil
}
