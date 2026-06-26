package chat_service

import (
	"context"
	"discord/internal/core/domain"
	"fmt"
)

func (s *ChatService) GetUserChats(ctx context.Context, user_id int64) ([]domain.Chat, error) {
	chats, err := s.chatRepository.GetUserChats(ctx, user_id)
	if err != nil {
		return nil, fmt.Errorf("get user chats: %w", err)
	}
	return chats, nil
}
