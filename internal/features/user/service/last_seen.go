package users_service

import (
	"context"
	"time"
)

func (s *UsersService) SaveLastSeen(ctx context.Context, userID int64, lastSeen time.Time) error {
	return s.usersRepository.SaveLastSeen(ctx, userID, lastSeen)
}

func (s *UsersService) GetLastSeen(ctx context.Context, userID int64) (*time.Time, error) {
	return s.usersRepository.GetLastSeen(ctx, userID)
}
