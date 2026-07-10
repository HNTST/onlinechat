package users_service

import (
	"context"
	"discord/internal/core/domain"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func (s *UsersService) SaveUser(ctx context.Context, user domain.User) (domain.User, error) {
	if err := user.Validate(); err != nil {
		return domain.User{}, fmt.Errorf("ошибка валидации домена пользователя: %w", err)
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, fmt.Errorf("хэширование пароля: %w", err)
	}
	user.Password = string(hashedPassword)
	user, err = s.usersRepository.SaveUser(ctx, user)
	if err != nil {
		return domain.User{}, fmt.Errorf("сохранение пользователя: %w", err)
	}
	return user, nil
}
