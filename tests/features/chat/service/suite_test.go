package chat_service_test

import (
	"context"
	"discord/internal/core/domain"

	"github.com/stretchr/testify/mock"
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
