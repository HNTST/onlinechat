package chat_service

import (
	"context"
	"fmt"
)

func (s *ChatService) GetUserChatIDs(ctx context.Context, userID int64) ([]int64, error) {
	ids, err := s.chatRepository.GetUserChatIDs(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user chat ids: %w", err)
	}
	return ids, nil
}
