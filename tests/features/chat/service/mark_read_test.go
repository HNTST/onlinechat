package chat_service_test

import (
	"context"
	"errors"
	"testing"

	chat_service "discord/internal/features/chat/service"

	"github.com/stretchr/testify/require"
)

func TestMarkMessagesAsRead_Success(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	mockRepo.On("MarkMessagesAsRead", ctx, int64(10), int64(42), int64(100)).Return(nil)

	err := svc.MarkMessagesAsRead(ctx, 10, 42, 100)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestMarkMessagesAsRead_Error(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	mockRepo.On("MarkMessagesAsRead", ctx, int64(10), int64(42), int64(100)).Return(errors.New("db error"))

	err := svc.MarkMessagesAsRead(ctx, 10, 42, 100)

	require.Error(t, err)
	mockRepo.AssertExpectations(t)
}
