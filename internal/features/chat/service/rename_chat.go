package chat_service

import (
	"context"
	core_errors "discord/internal/core/errors"
	"fmt"
)

func (s *ChatService) RenameChat(
	ctx context.Context,
	chatID int64,
	userID int64,
	newName string,
) error {
	chat, err := s.chatRepository.GetChatByID(ctx, userID, chatID)
	if err != nil {
		return fmt.Errorf("get chat by id: %w", err)
	}
	if !chat.IsGroup {
		return fmt.Errorf("chat is not a group: %w", core_errors.ErrInvalidArgument)
	}

	err = s.chatRepository.UpdateChatName(ctx, chatID, newName)
	if err != nil {
		return fmt.Errorf("update chat name: %w", err)
	}
	return nil
}
