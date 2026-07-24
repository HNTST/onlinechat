package chat_service

import (
	"context"
	"time"
)

func (s *ChatService) SaveLastSeen(ctx context.Context, userID int64, lastSeen time.Time) error {
	return s.chatRepository.SaveLastSeen(ctx, userID, lastSeen)
}

func (s *ChatService) GetLastSeen(ctx context.Context, userID int64) (*time.Time, error) {
	return s.chatRepository.GetLastSeen(ctx, userID)
}
