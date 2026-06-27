package users_service_test

import (
	"context"
	"errors"
	"testing"

	users_service "discord/internal/features/user/service"

	"github.com/stretchr/testify/require"
)

func TestGetUserID_Success(t *testing.T) {
	mockRepo := new(mockUserRepo)
	svc := users_service.NewUsersService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetUserID", ctx, "alice").Return(int64(42), nil)

	id, err := svc.GetUserID(ctx, "alice")

	require.NoError(t, err)
	require.Equal(t, int64(42), id)
	mockRepo.AssertExpectations(t)
}

func TestGetUserID_NotFound(t *testing.T) {
	mockRepo := new(mockUserRepo)
	svc := users_service.NewUsersService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetUserID", ctx, "nonexistent").Return(int64(0), errors.New("user not found"))

	_, err := svc.GetUserID(ctx, "nonexistent")

	require.Error(t, err)
	mockRepo.AssertExpectations(t)
}
