package chat_service

import (
	"context"
	"fmt"
)

func (s *ChatService) EditMessage(ctx context.Context, chatID, messageID, userID int64, text string) error {
	if _, err := s.chatRepository.GetChatByID(ctx, userID, chatID); err != nil {
		return fmt.Errorf("check chat membership: %w", err)
	}
	err := s.chatRepository.EditMessage(ctx, chatID, messageID, userID, text)
	if err != nil {
		return fmt.Errorf("edit message: %w", err)
	}
	return nil
}
