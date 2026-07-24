package chats_postgres_repository

import (
	"context"
	"fmt"
	"time"
)

func (r *ChatRepository) SaveLastSeen(ctx context.Context, userID int64, lastSeen time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `UPDATE users SET last_seen_at = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, lastSeen, userID)
	if err != nil {
		return fmt.Errorf("save last seen: %w", err)
	}
	return nil
}

func (r *ChatRepository) GetLastSeen(ctx context.Context, userID int64) (*time.Time, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `SELECT last_seen_at FROM users WHERE id = $1`
	var lastSeen *time.Time
	err := r.pool.QueryRow(ctx, query, userID).Scan(&lastSeen)
	if err != nil {
		return nil, fmt.Errorf("get last seen: %w", err)
	}
	return lastSeen, nil
}
