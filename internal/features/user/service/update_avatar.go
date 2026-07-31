package users_service

import (
	"context"
	"fmt"
)

func (s *UsersService) UpdateAvatar(ctx context.Context, userID, fileID int64) error {
	if err := s.usersRepository.UpdateAvatar(ctx, userID, fileID); err != nil {
		return fmt.Errorf("update user avatar: %w", err)
	}
	return nil
}
