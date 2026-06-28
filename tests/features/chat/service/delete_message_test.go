package chat_service_test

import (
	"context"
	"errors"
	"testing"

	chat_service "discord/internal/features/chat/service"

	"github.com/stretchr/testify/require"
)

func TestSoftDeleteMessage_Success(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	mockRepo.On("SoftDeleteMessage", ctx, int64(1)).Return(nil)

	err := svc.SoftDeleteMessage(ctx, 1)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestSoftDeleteMessage_Error(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	mockRepo.On("SoftDeleteMessage", ctx, int64(99)).Return(errors.New("db error"))

	err := svc.SoftDeleteMessage(ctx, 99)

	require.Error(t, err)
	mockRepo.AssertExpectations(t)
}
