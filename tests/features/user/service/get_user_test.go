package users_service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"discord/internal/core/domain"
	users_service "discord/internal/features/user/service"

	"github.com/stretchr/testify/require"
)

func TestGetUser_Success(t *testing.T) {
	mockRepo := new(mockUserRepo)
	svc := users_service.NewUsersService(mockRepo)
	ctx := context.Background()

	userID := 42
	expected := domain.NewUser(42, "charlie", "-1", time.Now())

	mockRepo.On("GetUser", ctx, int64(42)).Return(expected, nil)

	user, err := svc.GetUser(ctx, &userID)

	require.NoError(t, err)
	require.Equal(t, "charlie", user.Name)
	mockRepo.AssertExpectations(t)
}

func TestGetUser_SuccessWithAvatarID(t *testing.T) {
	mockRepo := new(mockUserRepo)
	svc := users_service.NewUsersService(mockRepo)
	ctx := context.Background()

	avatarID := int64(7)
	userID := 42
	expected := domain.NewUserWithAvatar(42, "charlie", "-1", time.Now(), &avatarID)

	mockRepo.On("GetUser", ctx, int64(42)).Return(expected, nil)

	user, err := svc.GetUser(ctx, &userID)

	require.NoError(t, err)
	require.NotNil(t, user.AvatarID)
	require.Equal(t, int64(7), *user.AvatarID)
	mockRepo.AssertExpectations(t)
}

func TestGetUser_InvalidArgument(t *testing.T) {
	mockRepo := new(mockUserRepo)
	svc := users_service.NewUsersService(mockRepo)
	ctx := context.Background()

	negID := -5

	_, err := svc.GetUser(ctx, &negID)

	require.Error(t, err)
	mockRepo.AssertNotCalled(t, "GetUser")
}

func TestGetUser_RepoError(t *testing.T) {
	mockRepo := new(mockUserRepo)
	svc := users_service.NewUsersService(mockRepo)
	ctx := context.Background()

	userID := 99

	mockRepo.On("GetUser", ctx, int64(99)).Return(domain.User{}, errors.New("not found"))

	_, err := svc.GetUser(ctx, &userID)

	require.Error(t, err)
	mockRepo.AssertExpectations(t)
}
