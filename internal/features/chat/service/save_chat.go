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
	user_id,
	target_user_id int64,
) (int64, error) {
	var (
		chat_id int64
		err     error
	)

	err = s.chatRepository.Transaction(
		ctx,
		func(txCtx context.Context,
		) error {

			chat_id, err = s.chatRepository.CheckExistsChat(
				txCtx,
				user_id,
				target_user_id,
			)
			if err != nil {
				if errors.Is(err, core_errors.ErrChatNotExists) {
					err = nil
					chat_id, err = s.chatRepository.CreateNewChat(
						txCtx,
						name,
					)
					if err != nil {
						return fmt.Errorf("create chat: %w", err)
					}
				}
				return fmt.Errorf("create chat: %w", err)
			}
			err = s.chatRepository.AddUserToChat(
				txCtx,
				chat_id,
				[]int64{user_id, target_user_id},
			)
			if err != nil {
				return fmt.Errorf("add users to chat: %w", err)
			}

			return nil
		})

	if err != nil {
		return -1, fmt.Errorf("save chat or add users to chat failed: %w", err)
	}
	return chat_id, nil
}
