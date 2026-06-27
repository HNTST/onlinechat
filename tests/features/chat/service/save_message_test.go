package chat_service_test

import (
	"context"
	"errors"
	"testing"

	"discord/internal/core/domain"
	chat_service "discord/internal/features/chat/service"

	"github.com/stretchr/testify/require"
)

func TestSaveMessage_Success(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	msg := domain.Messages{ChatId: 1, SenderId: 10, Message: "hello"}
	mockRepo.On("SaveMessage", ctx, msg).Return(nil)

	err := svc.SaveMessage(ctx, msg)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestSaveMessage_Error(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	msg := domain.Messages{ChatId: 1, SenderId: 10, Message: "hello"}
	mockRepo.On("SaveMessage", ctx, msg).Return(errors.New("insert error"))

	err := svc.SaveMessage(ctx, msg)

	require.Error(t, err)
	mockRepo.AssertExpectations(t)
}
