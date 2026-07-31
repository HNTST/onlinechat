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

func TestEditMessage_Success(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetChatByID", ctx, int64(1), int64(5)).Return(domain.Chat{Id: 5, IsGroup: true}, nil)
	mockRepo.On("EditMessage", ctx, int64(5), int64(1), int64(1), "new text").Return(nil)

	err := svc.EditMessage(ctx, 5, 1, 1, "new text")

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestEditMessage_Error(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetChatByID", ctx, int64(1), int64(5)).Return(domain.Chat{Id: 5, IsGroup: true}, nil)
	mockRepo.On("EditMessage", ctx, int64(5), int64(99), int64(1), "text").Return(errors.New("db error"))

	err := svc.EditMessage(ctx, 5, 99, 1, "text")

	require.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestEditMessage_UserNotMember(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetChatByID", ctx, int64(2), int64(5)).
		Return(domain.Chat{}, core_errors.ErrChatNotExists)

	err := svc.EditMessage(ctx, 5, 1, 2, "text")

	require.Error(t, err)
	require.ErrorIs(t, err, core_errors.ErrChatNotExists)
	mockRepo.AssertNotCalled(t, "EditMessage")
}
