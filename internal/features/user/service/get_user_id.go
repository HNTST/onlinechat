package users_service

import (
	"context"
	"fmt"
)

func (s *UsersService) GetUserID(ctx context.Context, username string) (int64, error) {
	id, err := s.usersRepository.GetUserID(ctx, username)
	if err != nil {
		return 0, fmt.Errorf("get user id from repository: %w", err)
	}
	return id, nil
}
