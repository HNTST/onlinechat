package domain

import "time"

const SessionTTL = 24 * time.Hour

type Session struct {
	ID        int64
	UserID    int64
	Token     string
	ExpiresAt time.Time
}
