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
	mockRepo.On("SaveMessage", ctx, msg).Return(int64(42), nil)

	id, err := svc.SaveMessage(ctx, msg)

	require.NoError(t, err)
	require.Equal(t, int64(42), id)
	mockRepo.AssertExpectations(t)
}

func TestSaveMessage_Error(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	msg := domain.Messages{ChatId: 1, SenderId: 10, Message: "hello"}
	mockRepo.On("SaveMessage", ctx, msg).Return(int64(0), errors.New("insert error"))

	id, err := svc.SaveMessage(ctx, msg)

	require.Error(t, err)
	require.Equal(t, int64(0), id)
	mockRepo.AssertExpectations(t)
}
