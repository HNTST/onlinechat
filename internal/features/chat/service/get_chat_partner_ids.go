package chat_service

import (
	"context"
	"fmt"
)

func (s *ChatService) GetChatPartnerIDs(ctx context.Context, userID int64) ([]int64, error) {
	partners, err := s.chatRepository.GetChatPartnerIDs(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get chat partner ids: %w", err)
	}
	return partners, nil
}
