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

type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) SaveUser(ctx context.Context, user domain.User) (domain.User, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(domain.User), args.Error(1)
}

func (m *mockUserRepo) GetUser(ctx context.Context, userID int64) (domain.User, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(domain.User), args.Error(1)
}

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
