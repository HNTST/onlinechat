package chat_service

import (
	"context"
	"fmt"
)

func (s *ChatService) SoftDeleteMessage(ctx context.Context, chatID, messageID int64) error {
	err := s.chatRepository.SoftDeleteMessage(ctx, chatID, messageID)
	if err != nil {
		return fmt.Errorf("soft delete message: %w", err)
	}
	return nil
}
