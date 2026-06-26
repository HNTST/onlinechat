package chat_service

import (
	"context"
	"discord/internal/core/domain"
	"fmt"
)

func (s *ChatService) GetChatMessages(ctx context.Context, chat_id int64) ([]domain.Messages, error) {
	messages, err := s.chatRepository.GetChatMessages(ctx, chat_id)
	if err != nil {
		return nil, fmt.Errorf("get chat messages: %w", err)
	}
	return messages, nil
}
