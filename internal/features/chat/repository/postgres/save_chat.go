package chats_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (r *ChatRepository) SaveChat(
	ctx context.Context,
	name string,
	user_id, target_user_id int64,
) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return -1, fmt.Errorf("tx create: %w", err)
	}
	defer tx.Rollback(ctx)
	// проверить, есть ли уже созданный чат
	existing_chat_id, err := r.CheckExistsChat(tx, ctx, user_id, target_user_id)
	if err != nil {
		return -1, fmt.Errorf("chect existing chat: %w", err)
	}
	if existing_chat_id > 0 {
		return existing_chat_id, nil
	}
	// если старого нет то создать
	chat_id, err := r.CreateNewChat(tx, ctx, name)
	if err != nil {
		return -1, fmt.Errorf("create new chat: %w", err)
	}

	err = r.AddUserToChat(ctx, chat_id, user_id)
	if err != nil {
		return -1, fmt.Errorf("add first user: %w", err)
	}
	err = r.AddUserToChat(ctx, chat_id, target_user_id)
	if err != nil {
		return -1, fmt.Errorf("add second user: %w", err)
	}
	err = tx.Commit(ctx)
	if err != nil {
		return -1, fmt.Errorf("commit: %w", err)
	}
	return chat_id, nil

}
func (r *ChatRepository) CheckExistsChat(
	tx pgx.Tx,
	ctx context.Context,
	user_id, target_user_id int64,
) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		select chat_id
		from chat_members
		where user_id in ($1, $2)
		group by chat_id
		having count(dstinct user_id) = 2
	`
	var chat_id int64

	row := tx.QueryRow(ctx, query, user_id, target_user_id)
	err := row.Scan(&chat_id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return -1, fmt.Errorf("get rows: no rows: %w", err)
		}
		return -1, fmt.Errorf("get rows: %w", err)
	}
	return chat_id, nil

}

func (r *ChatRepository) CreateNewChat(
	tx pgx.Tx,
	ctx context.Context,
	chat_name string,
) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		insert into chats (name)
		values($1)
		returning id
	`
	var chat_id int64
	row := tx.QueryRow(ctx, query, chat_name)
	err := row.Scan(chat_id)
	if err != nil {
		return -1, fmt.Errorf("scan error: %w", err)
	}
	return chat_id, nil
}
