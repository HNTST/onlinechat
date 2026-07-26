package chat_service

import (
	"context"
	core_errors "discord/internal/core/errors"
	"fmt"
)

func (s *ChatService) DeleteChat(
	ctx context.Context,
	chatID int64,
	userID int64,
) error {
	chat, err := s.chatRepository.GetChatByID(ctx, userID, chatID)
	if err != nil {
		return fmt.Errorf("get chat by id: %w", err)
	}
	if !chat.IsGroup {
		return fmt.Errorf("chat is not a group: %w", core_errors.ErrInvalidArgument)
	}

	err = s.chatRepository.DeleteChat(ctx, chatID)
	if err != nil {
		return fmt.Errorf("delete chat: %w", err)
	}
	return nil
}
