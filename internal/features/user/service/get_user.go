package users_service

import (
	"context"
	"discord/internal/core/domain"
	core_errors "discord/internal/core/errors"
	"fmt"
)

func (s *UsersService) GetUser(ctx context.Context, user_id *int) (domain.User, error) {
	if user_id != nil && *user_id < 0 {
		return domain.User{}, fmt.Errorf("user_id must be non negative: %w", core_errors.ErrInvalidArgument)
	}
	user, err := s.usersRepository.GetUser(ctx, int64(*user_id))
	if err != nil {
		return domain.User{}, fmt.Errorf("get user from repository: %w", err)
	}
	return user, nil
}
