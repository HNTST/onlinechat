package chat_service_test

import (
	"context"
	"errors"
	"testing"

	chat_service "discord/internal/features/chat/service"
	core_errors "discord/internal/core/errors"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSaveChat_CreatesNewChat(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	mockRepo.On("Transaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(nil)
	mockRepo.On("CheckExistsChat", ctx, int64(1), int64(2)).
		Return(int64(0), core_errors.ErrChatNotExists)
	mockRepo.On("CreateNewChat", ctx, "new-chat", false).
		Return(int64(42), nil)
	mockRepo.On("AddUserToChat", ctx, int64(42), []int64{1, 2}).
		Return(nil)

	chatID, err := svc.SaveChat(ctx, "new-chat", 1, 2, false)

	require.NoError(t, err)
	require.Equal(t, int64(42), chatID)
	mockRepo.AssertExpectations(t)
}

func TestSaveChat_ReturnsExisting(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	mockRepo.On("Transaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(nil)
	mockRepo.On("CheckExistsChat", ctx, int64(1), int64(2)).
		Return(int64(42), nil)
	mockRepo.On("AddUserToChat", ctx, int64(42), []int64{1, 2}).
		Return(nil)

	chatID, err := svc.SaveChat(ctx, "existing-chat", 1, 2, false)

	require.NoError(t, err)
	require.Equal(t, int64(42), chatID)
	mockRepo.AssertExpectations(t)
}

func TestSaveChat_CheckExistsError(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	mockRepo.On("Transaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(nil)
	mockRepo.On("CheckExistsChat", ctx, int64(1), int64(2)).
		Return(int64(0), errors.New("db error"))

	_, err := svc.SaveChat(ctx, "fail-chat", 1, 2, false)

	require.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestSaveChat_CreateNewChatError(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	mockRepo.On("Transaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(nil)
	mockRepo.On("CheckExistsChat", ctx, int64(1), int64(2)).
		Return(int64(0), core_errors.ErrChatNotExists)
	mockRepo.On("CreateNewChat", ctx, "broken-chat", false).
		Return(int64(0), errors.New("insert failed"))

	_, err := svc.SaveChat(ctx, "broken-chat", 1, 2, false)

	require.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestSaveChat_AddUserToChatError(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	mockRepo.On("Transaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(nil)
	mockRepo.On("CheckExistsChat", ctx, int64(1), int64(2)).
		Return(int64(42), nil)
	mockRepo.On("AddUserToChat", ctx, int64(42), []int64{1, 2}).
		Return(errors.New("add user failed"))

	_, err := svc.SaveChat(ctx, "fail-add", 1, 2, false)

	require.Error(t, err)
	mockRepo.AssertExpectations(t)
}
