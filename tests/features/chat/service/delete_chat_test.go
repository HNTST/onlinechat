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

func TestDeleteChat_Success(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetChatByID", ctx, int64(1), int64(42)).
		Return(domain.Chat{Id: 42, Name: "group-chat", IsGroup: true}, nil)
	mockRepo.On("DeleteChat", ctx, int64(42)).
		Return(nil)

	err := svc.DeleteChat(ctx, 42, 1)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDeleteChat_NotGroup(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetChatByID", ctx, int64(1), int64(42)).
		Return(domain.Chat{Id: 42, Name: "private-chat", IsGroup: false}, nil)

	err := svc.DeleteChat(ctx, 42, 1)

	require.Error(t, err)
	require.True(t, errors.Is(err, core_errors.ErrInvalidArgument))
	mockRepo.AssertExpectations(t)
}

func TestDeleteChat_UserNotMember(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetChatByID", ctx, int64(2), int64(42)).
		Return(domain.Chat{}, core_errors.ErrChatNotExists)

	err := svc.DeleteChat(ctx, 42, 2)

	require.Error(t, err)
	require.True(t, errors.Is(err, core_errors.ErrChatNotExists))
	mockRepo.AssertExpectations(t)
}

func TestDeleteChat_GetChatByIDError(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetChatByID", ctx, int64(1), int64(42)).
		Return(domain.Chat{}, errors.New("db error"))

	err := svc.DeleteChat(ctx, 42, 1)

	require.Error(t, err)
	mockRepo.AssertExpectations(t)
}
