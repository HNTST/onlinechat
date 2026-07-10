package auth_postgres_repository

import "time"

type SessionModel struct {
	ID        int64
	UserID    int64
	Token     string
	ExpiresAt time.Time
}
