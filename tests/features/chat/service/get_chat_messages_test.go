package chat_service_test

import (
	"context"
	"errors"
	"testing"

	"discord/internal/core/domain"
	chat_service "discord/internal/features/chat/service"

	"github.com/stretchr/testify/require"
)

func TestGetChatMessages_Success(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	expected := []domain.Messages{
		{ChatId: 1, SenderId: 10, Message: "hello"},
		{ChatId: 1, SenderId: 11, Message: "world"},
	}
	mockRepo.On("GetChatMessages", ctx, int64(1)).Return(expected, nil)

	messages, err := svc.GetChatMessages(ctx, 1)

	require.NoError(t, err)
	require.Len(t, messages, 2)
	require.Equal(t, "hello", messages[0].Message)
	mockRepo.AssertExpectations(t)
}

func TestGetChatMessages_Error(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetChatMessages", ctx, int64(99)).Return([]domain.Messages{}, errors.New("db error"))

	_, err := svc.GetChatMessages(ctx, 99)

	require.Error(t, err)
	mockRepo.AssertExpectations(t)
}
