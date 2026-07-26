package users_postgres_repository

import (
	"context"
	"discord/internal/db"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func (r *UserRepository) SaveLastSeen(ctx context.Context, userID int64, lastSeen time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	q := r.pool.GetQueries(ctx)
	err := q.UpdateUserLastSeen(ctx, db.UpdateUserLastSeenParams{
		LastSeenAt: pgtype.Timestamptz{Time: lastSeen, Valid: true},
		ID:         userID,
	})
	if err != nil {
		return fmt.Errorf("save last seen: %w", err)
	}
	return nil
}

func (r *UserRepository) GetLastSeen(ctx context.Context, userID int64) (*time.Time, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	q := r.pool.GetQueries(ctx)
	tz, err := q.GetUserLastSeenByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get last seen: %w", err)
	}

	var lastSeen *time.Time
	if tz.Valid {
		t := tz.Time
		lastSeen = &t
	}
	return lastSeen, nil
}
