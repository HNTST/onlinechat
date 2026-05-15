package chats_postgres_repository

import (
	"context"
	core_errors "discord/internal/core/errors"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// func (r *ChatRepository) SaveChat(
// 	ctx context.Context,
// 	name string,
// 	user_id, target_user_id int64,
// ) (int64, error) {
// 	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
// 	defer cancel()

// 	tx, err := r.pool.Begin(ctx)
// 	if err != nil {
// 		return -1, fmt.Errorf("tx create: %w", err)
// 	}
// 	defer tx.Rollback(ctx)
// 	// проверить, есть ли уже созданный чат
// 	existing_chat_id, err := r.CheckExistsChat(tx, ctx, user_id, target_user_id)
// 	if err != nil {
// 		return -1, fmt.Errorf("chect existing chat: %w", err)
// 	}
// 	if existing_chat_id > 0 {
// 		return existing_chat_id, nil
// 	}
// 	// если старого нет то создать
// 	chat_id, err := r.CreateNewChat(tx, ctx, name)
// 	if err != nil {
// 		return -1, fmt.Errorf("create new chat: %w", err)
// 	}

// 	err = r.AddUserToChat(ctx, chat_id, []int64{user_id, target_user_id})
// 	if err != nil {
// 		return -1, fmt.Errorf("add users: %w", err)
// 	}
// 	err = tx.Commit(ctx)
// 	if err != nil {
// 		return -1, fmt.Errorf("commit: %w", err)
// 	}
// 	return chat_id, nil

// }
func (r *ChatRepository) CheckExistsChat(
	ctx context.Context,
	user_id, target_user_id int64,
) (int64, error) {
	var (
		err     error
		chat_id int64

		row pgx.Row
	)

	query := `
		select chat_id
		from chat_members
		where user_id in ($1, $2)
		group by chat_id
		having count(distinct user_id) = 2
	`

	tx := r.pool.GetTx(ctx)

	if tx != nil {
		row = tx.QueryRow(ctx, query, user_id, target_user_id)
	} else {
		ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
		defer cancel()
		row = r.pool.QueryRow(ctx, query, user_id, target_user_id)
	}

	err = row.Scan(&chat_id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return -1, fmt.Errorf("get rows: no rows: %w", core_errors.ErrChatNotExists)
		}
		return -1, fmt.Errorf("get rows: %w", err)
	}
	return chat_id, nil

}
