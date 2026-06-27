package chat_service_test

import (
	"context"
	"errors"
	"testing"

	"discord/internal/core/domain"
	chat_service "discord/internal/features/chat/service"

	"github.com/stretchr/testify/require"
)

func TestGetUserChats_Success(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	expected := []domain.Chat{
		{Id: 1, Name: "chat-1"},
		{Id: 2, Name: "chat-2"},
	}
	mockRepo.On("GetUserChats", ctx, int64(42)).Return(expected, nil)

	chats, err := svc.GetUserChats(ctx, 42)

	require.NoError(t, err)
	require.Len(t, chats, 2)
	require.Equal(t, int64(1), chats[0].Id)
	mockRepo.AssertExpectations(t)
}

func TestGetUserChats_Error(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetUserChats", ctx, int64(99)).Return([]domain.Chat{}, errors.New("db error"))

	_, err := svc.GetUserChats(ctx, 99)

	require.Error(t, err)
	mockRepo.AssertExpectations(t)
}
