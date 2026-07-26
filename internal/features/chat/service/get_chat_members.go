package chat_service

import (
	"context"
	"fmt"
)

func (s *ChatService) GetChatMembers(ctx context.Context, chatID int64) ([]int64, error) {
	members, err := s.chatRepository.GetChatMembers(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("get chat members: %w", err)
	}
	return members, nil
}
