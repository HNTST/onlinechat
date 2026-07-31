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

func TestSoftDeleteMessage_Success(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetChatByID", ctx, int64(1), int64(10)).Return(domain.Chat{Id: 10, IsGroup: true}, nil)
	mockRepo.On("SoftDeleteMessage", ctx, int64(10), int64(1), int64(1)).Return(nil)

	err := svc.SoftDeleteMessage(ctx, 10, 1, 1)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestSoftDeleteMessage_Error(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetChatByID", ctx, int64(1), int64(10)).Return(domain.Chat{Id: 10, IsGroup: true}, nil)
	mockRepo.On("SoftDeleteMessage", ctx, int64(10), int64(99), int64(1)).Return(errors.New("db error"))

	err := svc.SoftDeleteMessage(ctx, 10, 99, 1)

	require.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestSoftDeleteMessage_UserNotMember(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetChatByID", ctx, int64(2), int64(10)).
		Return(domain.Chat{}, core_errors.ErrChatNotExists)

	err := svc.SoftDeleteMessage(ctx, 10, 1, 2)

	require.Error(t, err)
	require.ErrorIs(t, err, core_errors.ErrChatNotExists)
	mockRepo.AssertNotCalled(t, "SoftDeleteMessage")
}
