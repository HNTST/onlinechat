package chat_service_test

import (
	"context"
	"errors"
	"testing"

	chat_service "discord/internal/features/chat/service"

	"github.com/stretchr/testify/require"
)

func TestAddUserToChat_Success(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	mockRepo.On("AddUserToChat", ctx, int64(10), []int64{3, 4}).
		Return(nil)

	err := svc.AddUserToChat(ctx, 10, []int64{3, 4})

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAddUserToChat_Error(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	mockRepo.On("AddUserToChat", ctx, int64(10), []int64{3}).
		Return(errors.New("repo error"))

	err := svc.AddUserToChat(ctx, 10, []int64{3})

	require.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAddUserToChat_EmptyUsers(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	mockRepo.On("AddUserToChat", ctx, int64(10), []int64{}).
		Return(nil)

	err := svc.AddUserToChat(ctx, 10, []int64{})

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
