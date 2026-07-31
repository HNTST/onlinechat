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

func (m *mockChatRepo) CreateNewChat(ctx context.Context, chatName string, isGroup bool) (int64, error) {
	args := m.Called(ctx, chatName, isGroup)
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

func (m *mockChatRepo) SaveMessage(ctx context.Context, message domain.Messages) (int64, error) {
	args := m.Called(ctx, message)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockChatRepo) EditMessage(ctx context.Context, chatID, messageID, userID int64, text string) error {
	args := m.Called(ctx, chatID, messageID, userID, text)
	return args.Error(0)
}

func (m *mockChatRepo) SoftDeleteMessage(ctx context.Context, chatID, messageID, userID int64) error {
	args := m.Called(ctx, chatID, messageID, userID)
	return args.Error(0)
}

func (m *mockChatRepo) MarkMessagesAsRead(ctx context.Context, chatID, userID, lastReadMessageID int64) error {
	args := m.Called(ctx, chatID, userID, lastReadMessageID)
	return args.Error(0)
}

func (m *mockChatRepo) GetChatMessagesPaginated(ctx context.Context, chatID int64, lastMessageID *int64, limit int) ([]domain.Messages, error) {
	args := m.Called(ctx, chatID, lastMessageID, limit)
	return args.Get(0).([]domain.Messages), args.Error(1)
}

func (m *mockChatRepo) GetUserChatsExtended(ctx context.Context, userID int64) ([]domain.ChatInfo, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.ChatInfo), args.Error(1)
}

func (m *mockChatRepo) UpdateChatName(ctx context.Context, chatID int64, newName string) error {
	args := m.Called(ctx, chatID, newName)
	return args.Error(0)
}

func (m *mockChatRepo) DeleteChat(ctx context.Context, chatID int64) error {
	args := m.Called(ctx, chatID)
	return args.Error(0)
}

func (m *mockChatRepo) GetChatPartnerIDs(ctx context.Context, userID int64) ([]int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]int64), args.Error(1)
}

func (m *mockChatRepo) GetChatByID(ctx context.Context, userID int64, chatID int64) (domain.Chat, error) {
	args := m.Called(ctx, userID, chatID)
	return args.Get(0).(domain.Chat), args.Error(1)
}

func (m *mockChatRepo) GetUserChatIDs(ctx context.Context, userID int64) ([]int64, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]int64), args.Error(1)
}
