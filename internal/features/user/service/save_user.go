package users_service

import (
	"context"
	"discord/internal/core/domain"
	"fmt"
)

func (s *UsersService) SaveUser(ctx context.Context, user domain.User) (domain.User, error) {
	if err := user.Validate(); err != nil {
		return domain.User{}, fmt.Errorf("ошибка валидации домена пользователя: %w", err)
	}
	user, err := s.usersRepository.SaveUser(ctx, user)
	if err != nil {
		return domain.User{}, fmt.Errorf("сохранение пользователя: %w", err)
	}
	return user, nil
}
