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
	var chatID int64

	err := s.chatRepository.Transaction(
		ctx,
		func(txCtx context.Context) error {

			if !isGroup {
				existsID, checkErr := s.chatRepository.CheckExistsChat(
					txCtx,
					userID,
					targetUserID,
				)
				if checkErr != nil {
					if errors.Is(checkErr, core_errors.ErrChatNotExists) {
						chatID = 0
					} else {
						return fmt.Errorf("check existing chat: %w", checkErr)
					}
				} else {
					chatID = existsID
				}
			}

			if chatID == 0 {
				createdID, createErr := s.chatRepository.CreateNewChat(
					txCtx,
					name,
					isGroup,
				)
				if createErr != nil {
					return fmt.Errorf("create new chat: %w", createErr)
				}
				chatID = createdID
			}

			if addErr := s.chatRepository.AddUserToChat(
				txCtx,
				chatID,
				[]int64{userID, targetUserID},
			); addErr != nil {
				return fmt.Errorf("add users to chat: %w", addErr)
			}

			return nil
		})

	if err != nil {
		return 0, fmt.Errorf("save chat or add users to chat failed: %w", err)
	}
	return chatID, nil
}
