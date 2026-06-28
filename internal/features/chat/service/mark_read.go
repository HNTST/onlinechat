package chat_service

import (
	"context"
	"fmt"
)

func (s *ChatService) MarkMessagesAsRead(
	ctx context.Context,
	chatID int64,
	userID int64,
	lastReadMessageID int64,
) error {
	err := s.chatRepository.MarkMessagesAsRead(ctx, chatID, userID, lastReadMessageID)
	if err != nil {
		return fmt.Errorf("mark messages as read: %w", err)
	}
	return nil
}
