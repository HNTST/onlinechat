package chat_service

import (
	"context"
	"discord/internal/core/domain"
	"fmt"
)

func (s *ChatService) EditMessage(ctx context.Context, messageID int64, text string) error {
	msg := domain.Messages{Id: messageID, Message: text}
	err := s.chatRepository.EditMessage(ctx, msg.Id, msg.Message)
	if err != nil {
		return fmt.Errorf("edit message: %w", err)
	}
	return nil
}
