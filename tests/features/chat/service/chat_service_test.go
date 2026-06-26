package chat_service_test

import (
	"context"
	"discord/internal/core/domain"
	"errors"
	"testing"

	chat_service "discord/internal/features/chat/service"
	core_errors "discord/internal/core/errors"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockChatRepo struct {
	mock.Mock
}

func (m *mockChatRepo) Transaction(ctx context.Context, fn func(context.Context) error) error {
	m.Called(ctx, fn)
	return fn(ctx)
}

func (m *mockChatRepo) CheckExistsChat(ctx context.Context, userID, targetUserID int64) (int64, error) {
	args := m.Called(ctx, userID, targetUserID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockChatRepo) CreateNewChat(ctx context.Context, chatName string) (int64, error) {
	args := m.Called(ctx, chatName)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockChatRepo) AddUserToChat(ctx context.Context, chatID int64, usersIDs []int64) error {
	args := m.Called(ctx, chatID, usersIDs)
	return args.Error(0)
}

func (m *mockChatRepo) GetUserChats(ctx context.Context, user_id int64) ([]domain.Chat, error) {
	args := m.Called(ctx, user_id)
	return args.Get(0).([]domain.Chat), args.Error(1)
}

func (m *mockChatRepo) GetChatMessages(ctx context.Context, chat_id int64) ([]domain.Messages, error) {
	args := m.Called(ctx, chat_id)
	return args.Get(0).([]domain.Messages), args.Error(1)
}

func (m *mockChatRepo) GetChatMembers(ctx context.Context, chat_id int64) ([]int64, error) {
	args := m.Called(ctx, chat_id)
	return args.Get(0).([]int64), args.Error(1)
}

func (m *mockChatRepo) SaveMessage(ctx context.Context, message domain.Messages) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func TestSaveChat_CreatesNewChat(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	mockRepo.On("Transaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(nil)
	mockRepo.On("CheckExistsChat", ctx, int64(1), int64(2)).
		Return(int64(0), core_errors.ErrChatNotExists)
	mockRepo.On("CreateNewChat", ctx, "new-chat").
		Return(int64(42), nil)

	chatID, err := svc.SaveChat(ctx, "new-chat", 1, 2)

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

	chatID, err := svc.SaveChat(ctx, "existing-chat", 1, 2)

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

	_, err := svc.SaveChat(ctx, "fail-chat", 1, 2)

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
	mockRepo.On("CreateNewChat", ctx, "broken-chat").
		Return(int64(0), errors.New("insert failed"))

	_, err := svc.SaveChat(ctx, "broken-chat", 1, 2)

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

	_, err := svc.SaveChat(ctx, "fail-add", 1, 2)

	require.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAddUserToChat_Success(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	mockRepo.On("AddUserToChat", ctx, int64(10), []int64{3, 4}).
		Return(nil)

	err := svc.AddUserToChat(ctx, 10, []int64{3, 4})

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAddUserToChat_Error(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	mockRepo.On("AddUserToChat", ctx, int64(10), []int64{3}).
		Return(errors.New("repo error"))

	err := svc.AddUserToChat(ctx, 10, []int64{3})

	require.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestGetUserChats_Success(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	expected := []domain.Chat{
		{Id: 1, Name: "chat-1"},
		{Id: 2, Name: "chat-2"},
	}
	mockRepo.On("GetUserChats", ctx, int64(42)).Return(expected, nil)

	chats, err := svc.GetUserChats(ctx, 42)

	require.NoError(t, err)
	require.Len(t, chats, 2)
	require.Equal(t, int64(1), chats[0].Id)
	mockRepo.AssertExpectations(t)
}

func TestGetUserChats_Error(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetUserChats", ctx, int64(99)).Return([]domain.Chat{}, errors.New("db error"))

	_, err := svc.GetUserChats(ctx, 99)

	require.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestGetChatMessages_Success(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	expected := []domain.Messages{
		{ChatId: 1, SenderId: 10, Message: "hello"},
		{ChatId: 1, SenderId: 11, Message: "world"},
	}
	mockRepo.On("GetChatMessages", ctx, int64(1)).Return(expected, nil)

	messages, err := svc.GetChatMessages(ctx, 1)

	require.NoError(t, err)
	require.Len(t, messages, 2)
	require.Equal(t, "hello", messages[0].Message)
	mockRepo.AssertExpectations(t)
}

func TestGetChatMessages_Error(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetChatMessages", ctx, int64(99)).Return([]domain.Messages{}, errors.New("db error"))

	_, err := svc.GetChatMessages(ctx, 99)

	require.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestGetChatMembers_Success(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	expected := []int64{10, 20, 30}
	mockRepo.On("GetChatMembers", ctx, int64(1)).Return(expected, nil)

	members, err := svc.GetChatMembers(ctx, 1)

	require.NoError(t, err)
	require.Equal(t, []int64{10, 20, 30}, members)
	mockRepo.AssertExpectations(t)
}

func TestGetChatMembers_Error(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetChatMembers", ctx, int64(99)).Return([]int64{}, errors.New("db error"))

	_, err := svc.GetChatMembers(ctx, 99)

	require.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestSaveMessage_Success(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	msg := domain.Messages{ChatId: 1, SenderId: 10, Message: "hello"}
	mockRepo.On("SaveMessage", ctx, msg).Return(nil)

	err := svc.SaveMessage(ctx, msg)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestSaveMessage_Error(t *testing.T) {
	mockRepo := new(mockChatRepo)
	svc := chat_service.NewChatService(mockRepo)
	ctx := context.Background()

	msg := domain.Messages{ChatId: 1, SenderId: 10, Message: "hello"}
	mockRepo.On("SaveMessage", ctx, msg).Return(errors.New("insert error"))

	err := svc.SaveMessage(ctx, msg)

	require.Error(t, err)
	mockRepo.AssertExpectations(t)
}
