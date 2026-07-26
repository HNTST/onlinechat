package chat_service

import (
	"context"
	"discord/internal/core/domain"
	"fmt"
)

func (s *ChatService) GetUserChats(ctx context.Context, userID int64) ([]domain.Chat, error) {
	chats, err := s.chatRepository.GetUserChats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user chats: %w", err)
	}
	return chats, nil
}

func (s *ChatService) GetUserChatsExtended(ctx context.Context, userID int64) ([]domain.ChatInfo, error) {
	chats, err := s.chatRepository.GetUserChatsExtended(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user chats extended: %w", err)
	}
	return chats, nil
}
