package users_service

import (
	"context"
	"discord/internal/core/domain"
	core_errors "discord/internal/core/errors"
	"fmt"
)

func (s *UsersService) GetUser(ctx context.Context, userID *int) (domain.User, error) {
	if userID != nil && *userID < 0 {
		return domain.User{}, fmt.Errorf("userID must be non negative: %w", core_errors.ErrInvalidArgument)
	}
	user, err := s.usersRepository.GetUser(ctx, int64(*userID))
	if err != nil {
		return domain.User{}, fmt.Errorf("get user from repository: %w", err)
	}
	return user, nil
}
