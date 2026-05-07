package chat_service

import (
	"context"
	"fmt"
)

func (s *ChatService) SaveChat(ctx context.Context,
	name string,
	user_id,
	target_user_id int64,
) (int64, error) {
	chat_id, err := s.SaveChat(ctx, name, user_id, target_user_id)
	if err != nil {
		return -1, fmt.Errorf("save chat: %w", err)
	}
	return chat_id, err
}
