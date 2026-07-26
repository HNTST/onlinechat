package chat_service

import (
	"context"
	core_errors "discord/internal/core/errors"
	"errors"
	"fmt"
)

func (s *ChatService) SaveChat(
	ctx context.Context,
	name string,
	userID,
	targetUserID int64,
	isGroup bool,
) (int64, error) {
	var (
		chatID int64
		err    error
	)

	err = s.chatRepository.Transaction(
		ctx,
		func(txCtx context.Context,
		) error {

			if !isGroup {
				chatID, err = s.chatRepository.CheckExistsChat(
					txCtx,
					userID,
					targetUserID,
				)
				if err != nil {
					if !errors.Is(err, core_errors.ErrChatNotExists) {
						return fmt.Errorf("check existing chat: %w", err)
					}
				}
			}

			if chatID == 0 {
				chatID, err = s.chatRepository.CreateNewChat(
					txCtx,
					name,
					isGroup,
				)
				if err != nil {
					return fmt.Errorf("create new chat: %w", err)
				}
			}
			err = s.chatRepository.AddUserToChat(
				txCtx,
				chatID,
				[]int64{userID, targetUserID},
			)
			if err != nil {
				return fmt.Errorf("add users to chat: %w", err)
			}

			return nil
		})

	if err != nil {
		return 0, fmt.Errorf("save chat or add users to chat failed: %w", err)
	}
	return chatID, nil
}
