package chat_service

import (
	"context"
	"discord/internal/core/domain"
	"fmt"
)

func (s *ChatService) SaveMessage(ctx context.Context, message domain.Messages) (int64, error) {
	id, err := s.chatRepository.SaveMessage(ctx, message)
	if err != nil {
		return 0, fmt.Errorf("save message: %w", err)
	}
	return id, nil
}
