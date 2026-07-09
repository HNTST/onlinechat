package chat_service

import (
	"context"
	"discord/internal/core/domain"
	core_errors "discord/internal/core/errors"
	"fmt"
)

func (s *ChatService) DeleteChat(
	ctx context.Context,
	chatID int64,
	userID int64,
) error {
	isMember, err := s.chatRepository.IsUserInChat(ctx, userID, chatID)
	if err != nil {
		return fmt.Errorf("check user in chat: %w", err)
	}
	if !isMember {
		return fmt.Errorf("user is not a member: %w", core_errors.ErrChatNotExists)
	}

	chats, err := s.chatRepository.GetUserChats(ctx, userID)
	if err != nil {
		return fmt.Errorf("get chat info: %w", err)
	}

	var chat *domain.Chat
	for i := range chats {
		if chats[i].Id == chatID {
			chat = &chats[i]
			break
		}
	}
	if chat == nil {
		return fmt.Errorf("chat not found: %w", core_errors.ErrChatNotExists)
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
