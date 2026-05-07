package users_service

import (
	"context"
	"discord/internal/core/domain"
)

type UsersService struct {
	usersRepository UsersRepository
}

type UsersRepository interface {
	SaveUser(ctx context.Context, user domain.User) (domain.User, error)
	GetUser(ctx context.Context, user_id int64) (domain.User, error)
}

func NewUsersService(usersRepository UsersRepository) *UsersService {
	return &UsersService{
		usersRepository: usersRepository,
	}
}
