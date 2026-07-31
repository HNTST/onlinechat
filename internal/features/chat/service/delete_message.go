package chat_service

import (
	"context"
	"fmt"
)

func (s *ChatService) SoftDeleteMessage(ctx context.Context, chatID, messageID, userID int64) error {
	if _, err := s.chatRepository.GetChatByID(ctx, userID, chatID); err != nil {
		return fmt.Errorf("check chat membership: %w", err)
	}
	err := s.chatRepository.SoftDeleteMessage(ctx, chatID, messageID, userID)
	if err != nil {
		return fmt.Errorf("soft delete message: %w", err)
	}
	return nil
}
