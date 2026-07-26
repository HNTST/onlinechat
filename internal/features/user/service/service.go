package users_service

import (
	"context"
	"discord/internal/core/domain"
	"time"
)

type UsersService struct {
	usersRepository UsersRepository
}

type UsersRepository interface {
	SaveUser(ctx context.Context, user domain.User) (domain.User, error)
	GetUser(ctx context.Context, userID int64) (domain.User, error)
	GetUserID(ctx context.Context, username string) (int64, error)
	SaveLastSeen(ctx context.Context, userID int64, lastSeen time.Time) error
	GetLastSeen(ctx context.Context, userID int64) (*time.Time, error)
}

func NewUsersService(usersRepository UsersRepository) *UsersService {
	return &UsersService{
		usersRepository: usersRepository,
	}
}
