package chat_service

import (
	"context"
	"fmt"
)

func (s *ChatService) GetChatMembers(ctx context.Context, chat_id int64) ([]int64, error) {
	members, err := s.chatRepository.GetChatMembers(ctx, chat_id)
	if err != nil {
		return nil, fmt.Errorf("get chat members: %w", err)
	}
	return members, nil
}
