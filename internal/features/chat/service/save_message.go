package chat_service

import (
	"context"
	"discord/internal/core/domain"
	"fmt"
)

func (s *ChatService) SaveMessage(ctx context.Context, message domain.Messages) error {
	err := s.chatRepository.SaveMessage(ctx, message)
	if err != nil {
		return fmt.Errorf("save message: %w", err)
	}
	return nil
}
