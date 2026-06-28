package chat_service

import (
	"context"
	"fmt"
)

func (s *ChatService) SoftDeleteMessage(ctx context.Context, messageID int64) error {
	err := s.chatRepository.SoftDeleteMessage(ctx, messageID)
	if err != nil {
		return fmt.Errorf("soft delete message: %w", err)
	}
	return nil
}
