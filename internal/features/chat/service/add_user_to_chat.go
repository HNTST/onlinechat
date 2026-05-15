package chat_service

import (
	"context"
	"fmt"
)

func (s *ChatService) AddUserToChat(
	ctx context.Context,
	chat_id int64,
	user_id []int64,
) error {
	err := s.chatRepository.AddUserToChat(ctx, chat_id, user_id)
	if err != nil {
		return fmt.Errorf("add user to chat: %w", err)
	}
	return nil
}
