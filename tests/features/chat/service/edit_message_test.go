package chat_service_test

import (
	"context"
	"errors"
	"testing"

	chat_service "discord/internal/features/chat/service"

	"github.com/stretchr/testify/require"
)

func TestEditMessage_Success(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	mockRepo.On("EditMessage", ctx, int64(1), "new text").Return(nil)

	err := svc.EditMessage(ctx, 1, "new text")

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestEditMessage_Error(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	mockRepo.On("EditMessage", ctx, int64(99), "text").Return(errors.New("db error"))

	err := svc.EditMessage(ctx, 99, "text")

	require.Error(t, err)
	mockRepo.AssertExpectations(t)
}
