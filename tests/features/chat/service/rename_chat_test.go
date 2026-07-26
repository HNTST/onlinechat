package chat_service_test

import (
	"context"
	"discord/internal/core/domain"
	core_errors "discord/internal/core/errors"
	"errors"
	"testing"

	chat_service "discord/internal/features/chat/service"

	"github.com/stretchr/testify/require"
)

func TestRenameChat_Success(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetChatByID", ctx, int64(1), int64(42)).
		Return(domain.Chat{Id: 42, Name: "old-name", IsGroup: true}, nil)
	mockRepo.On("UpdateChatName", ctx, int64(42), "new-name").
		Return(nil)

	err := svc.RenameChat(ctx, 42, 1, "new-name")

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestRenameChat_NotGroup(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetChatByID", ctx, int64(1), int64(42)).
		Return(domain.Chat{Id: 42, Name: "private-chat", IsGroup: false}, nil)

	err := svc.RenameChat(ctx, 42, 1, "new-name")

	require.Error(t, err)
	require.True(t, errors.Is(err, core_errors.ErrInvalidArgument))
	mockRepo.AssertExpectations(t)
}

func TestRenameChat_UserNotMember(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetChatByID", ctx, int64(2), int64(42)).
		Return(domain.Chat{}, core_errors.ErrChatNotExists)

	err := svc.RenameChat(ctx, 42, 2, "new-name")

	require.Error(t, err)
	require.True(t, errors.Is(err, core_errors.ErrChatNotExists))
	mockRepo.AssertExpectations(t)
}

func TestRenameChat_ChatNotFound(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetChatByID", ctx, int64(1), int64(999)).
		Return(domain.Chat{}, core_errors.ErrChatNotExists)

	err := svc.RenameChat(ctx, 999, 1, "new-name")

	require.Error(t, err)
	require.True(t, errors.Is(err, core_errors.ErrChatNotExists))
	mockRepo.AssertExpectations(t)
}

func TestRenameChat_GetChatByIDError(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetChatByID", ctx, int64(1), int64(42)).
		Return(domain.Chat{}, errors.New("db error"))

	err := svc.RenameChat(ctx, 42, 1, "new-name")

	require.Error(t, err)
	mockRepo.AssertExpectations(t)
}
