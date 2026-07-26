package chat_service

import (
	"context"
	"fmt"
)

func (s *ChatService) AddUserToChat(
	ctx context.Context,
	chatID int64,
	userIDs []int64,
) error {
	err := s.chatRepository.AddUserToChat(ctx, chatID, userIDs)
	if err != nil {
		return fmt.Errorf("add user to chat: %w", err)
	}
	return nil
}
