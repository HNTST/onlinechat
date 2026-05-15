package chats_postgres_repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (r *ChatRepository) CreateNewChat(
	ctx context.Context,
	chat_name string,
) (int64, error) {
	var (
		err     error
		chat_id int64
		row     pgx.Row
	)
	query := `
		insert into chats (name)
		values($1)
		returning id
	`

	tx := r.pool.GetTx(ctx)

	if tx != nil {
		row = tx.QueryRow(ctx, query, chat_name)

	} else {
		ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
		defer cancel()
		row = r.pool.QueryRow(ctx, query, chat_name)
	}
	err = row.Scan(&chat_id)
	if err != nil {
		return -1, fmt.Errorf("scan error: %w", err)
	}
	return chat_id, nil
}
