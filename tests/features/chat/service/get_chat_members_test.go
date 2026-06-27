package chat_service_test

import (
	"context"
	"errors"
	"testing"

	chat_service "discord/internal/features/chat/service"

	"github.com/stretchr/testify/require"
)

func TestGetChatMembers_Success(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	expected := []int64{10, 20, 30}
	mockRepo.On("GetChatMembers", ctx, int64(1)).Return(expected, nil)

	members, err := svc.GetChatMembers(ctx, 1)

	require.NoError(t, err)
	require.Equal(t, []int64{10, 20, 30}, members)
	mockRepo.AssertExpectations(t)
}

func TestGetChatMembers_Error(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetChatMembers", ctx, int64(99)).Return([]int64{}, errors.New("db error"))

	_, err := svc.GetChatMembers(ctx, 99)

	require.Error(t, err)
	mockRepo.AssertExpectations(t)
}
