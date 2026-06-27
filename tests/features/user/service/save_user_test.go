package users_service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"discord/internal/core/domain"
	users_service "discord/internal/features/user/service"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSaveUser_Success(t *testing.T) {
	mockRepo := new(mockUserRepo)
	svc := users_service.NewUsersService(mockRepo)
	ctx := context.Background()

	input := domain.NewUserUninitialized("alice", "secret123")
	expected := domain.NewUser(42, input.Name, "-1", time.Now())

	mockRepo.On("SaveUser", ctx, mock.MatchedBy(func(u domain.User) bool {
		return u.Name == "alice"
	})).Return(expected, nil)

	user, err := svc.SaveUser(ctx, input)

	require.NoError(t, err)
	require.Equal(t, int64(42), user.Id)
	require.Equal(t, "alice", user.Name)
	mockRepo.AssertExpectations(t)
}

func TestSaveUser_ValidationError(t *testing.T) {
	mockRepo := new(mockUserRepo)
	svc := users_service.NewUsersService(mockRepo)
	ctx := context.Background()

	input := domain.NewUserUninitialized("ab", "secret123")

	_, err := svc.SaveUser(ctx, input)

	require.Error(t, err)
	mockRepo.AssertNotCalled(t, "SaveUser")
}

func TestSaveUser_RepoError(t *testing.T) {
	mockRepo := new(mockUserRepo)
	svc := users_service.NewUsersService(mockRepo)
	ctx := context.Background()

	input := domain.NewUserUninitialized("bob", "pass")

	mockRepo.On("SaveUser", ctx, mock.MatchedBy(func(u domain.User) bool {
		return u.Name == "bob"
	})).Return(domain.User{}, errors.New("duplicate key"))

	_, err := svc.SaveUser(ctx, input)

	require.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestSaveUser_SuccessWithAvatarID(t *testing.T) {
	mockRepo := new(mockUserRepo)
	svc := users_service.NewUsersService(mockRepo)
	ctx := context.Background()

	avatarID := int64(42)
	input := domain.NewUserUninitialized("alice", "secret123")
	expected := domain.NewUserWithAvatar(42, input.Name, "-1", time.Now(), &avatarID)

	mockRepo.On("SaveUser", ctx, mock.MatchedBy(func(u domain.User) bool {
		return u.Name == "alice"
	})).Return(expected, nil)

	user, err := svc.SaveUser(ctx, input)

	require.NoError(t, err)
	require.NotNil(t, user.AvatarID)
	require.Equal(t, int64(42), *user.AvatarID)
	mockRepo.AssertExpectations(t)
}
