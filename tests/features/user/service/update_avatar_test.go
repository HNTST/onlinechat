package users_service_test

import (
	"context"
	"errors"
	"testing"

	users_service "discord/internal/features/user/service"

	"github.com/stretchr/testify/require"
)

func TestUpdateAvatar_Success(t *testing.T) {
	mockRepo := new(mockUserRepo)
	svc := users_service.NewUsersService(mockRepo)
	ctx := context.Background()

	mockRepo.On("UpdateAvatar", ctx, int64(42), int64(7)).Return(nil)

	err := svc.UpdateAvatar(ctx, 42, 7)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUpdateAvatar_RepoError(t *testing.T) {
	mockRepo := new(mockUserRepo)
	svc := users_service.NewUsersService(mockRepo)
	ctx := context.Background()

	mockRepo.On("UpdateAvatar", ctx, int64(42), int64(7)).Return(errors.New("db error"))

	err := svc.UpdateAvatar(ctx, 42, 7)

	require.Error(t, err)
	mockRepo.AssertExpectations(t)
}
