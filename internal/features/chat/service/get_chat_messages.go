package chat_service

import (
	"context"
	"discord/internal/core/domain"
	"fmt"
)

func (s *ChatService) GetChatMessages(ctx context.Context, chatID int64) ([]domain.Messages, error) {
	messages, err := s.chatRepository.GetChatMessages(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("get chat messages: %w", err)
	}
	return messages, nil
}

func (s *ChatService) GetChatMessagesPaginated(
	ctx context.Context,
	chatID int64,
	lastMessageID *int64,
	limit int,
) ([]domain.Messages, error) {
	messages, err := s.chatRepository.GetChatMessagesPaginated(ctx, chatID, lastMessageID, limit)
	if err != nil {
		return nil, fmt.Errorf("get chat messages paginated: %w", err)
	}
	return messages, nil
}
