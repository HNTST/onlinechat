package chat_transport_http_test

import (
	"context"
	"discord/internal/core/domain"
	core_logger "discord/internal/core/logger"
	core_http_middleware "discord/internal/core/transport/middleware"
	core_http_server "discord/internal/core/transport/server"
	chat_transport_http "discord/internal/features/chat/transport/http"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockChatService struct {
	mock.Mock
}

func (m *mockChatService) SaveChat(ctx context.Context, name string, user_id, target_user_id int64) (int64, error) {
	args := m.Called(ctx, name, user_id, target_user_id)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockChatService) AddUserToChat(ctx context.Context, chat_id int64, user_id []int64) error {
	args := m.Called(ctx, chat_id, user_id)
	return args.Error(0)
}

func (m *mockChatService) GetUserChats(ctx context.Context, user_id int64) ([]domain.Chat, error) {
	args := m.Called(ctx, user_id)
	return args.Get(0).([]domain.Chat), args.Error(1)
}

func (m *mockChatService) GetChatMessages(ctx context.Context, chat_id int64) ([]domain.Messages, error) {
	args := m.Called(ctx, chat_id)
	return args.Get(0).([]domain.Messages), args.Error(1)
}

func (m *mockChatService) GetChatMembers(ctx context.Context, chat_id int64) ([]int64, error) {
	args := m.Called(ctx, chat_id)
	return args.Get(0).([]int64), args.Error(1)
}

func (m *mockChatService) SaveMessage(ctx context.Context, message domain.Messages) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *mockChatService) EditMessage(ctx context.Context, messageID int64, text string) error {
	args := m.Called(ctx, messageID, text)
	return args.Error(0)
}

func (m *mockChatService) SoftDeleteMessage(ctx context.Context, messageID int64) error {
	args := m.Called(ctx, messageID)
	return args.Error(0)
}

func (m *mockChatService) MarkMessagesAsRead(ctx context.Context, chatID, userID, lastReadMessageID int64) error {
	args := m.Called(ctx, chatID, userID, lastReadMessageID)
	return args.Error(0)
}

func (m *mockChatService) GetChatMessagesPaginated(ctx context.Context, chatID int64, lastMessageID *int64, limit int) ([]domain.Messages, error) {
	args := m.Called(ctx, chatID, lastMessageID, limit)
	return args.Get(0).([]domain.Messages), args.Error(1)
}

func (m *mockChatService) GetUserChatsExtended(ctx context.Context, userID int64) ([]domain.ChatInfo, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.ChatInfo), args.Error(1)
}

func setupChatServer(t *testing.T, svc *mockChatService) *httptest.Server {
	t.Helper()
	logger := core_logger.NewLogger()
	handler := chat_transport_http.NewChatHttpHandler(svc)
	apiRouter := core_http_server.NewApiVersionRouter(core_http_server.ApiVersion2)
	apiRouter.RegisterRoutes(handler.Routes()...)

	mux := http.NewServeMux()
	mux.Handle("/api/v2/", http.StripPrefix("/api/v2", apiRouter))

	wrapped := core_http_middleware.ChainMiddleware(
		mux,
		core_http_middleware.Logger(logger),
		core_http_middleware.Panic(),
	)
	return httptest.NewServer(wrapped)
}

func TestChatTransport_AddUserToChat_Success(t *testing.T) {
	mockSvc := new(mockChatService)
	srv := setupChatServer(t, mockSvc)
	defer srv.Close()

	mockSvc.On("AddUserToChat", mock.Anything, int64(42), []int64{1, 2}).Return(nil)

	bodyReader := strings.NewReader(`{"users_ids": [1, 2]}`)
	resp, err := http.Post(srv.URL+"/api/v2/chat/42/members", "application/json", bodyReader)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal(body, &result))
	require.Equal(t, float64(42), result["Chat_id"])
	mockSvc.AssertExpectations(t)
}

func TestChatTransport_AddUserToChat_InvalidPathParam(t *testing.T) {
	mockSvc := new(mockChatService)
	srv := setupChatServer(t, mockSvc)
	defer srv.Close()

	bodyReader := strings.NewReader(`{"users_ids": [1, 2]}`)
	resp, err := http.Post(srv.URL+"/api/v2/chat/abc/members", "application/json", bodyReader)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]string
	require.NoError(t, json.Unmarshal(body, &result))
	require.Contains(t, result["error"], "неверный аргумент")
	mockSvc.AssertNotCalled(t, "AddUserToChat")
}

func TestChatTransport_EditMessage_Success(t *testing.T) {
	mockSvc := new(mockChatService)
	srv := setupChatServer(t, mockSvc)
	defer srv.Close()

	mockSvc.On("EditMessage", mock.Anything, int64(1), "new text").Return(nil)

	bodyReader := strings.NewReader(`{"text": "new text"}`)
	req, err := http.NewRequest(http.MethodPut, srv.URL+"/api/v2/chats/5/messages/1", bodyReader)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	mockSvc.AssertExpectations(t)
}

func TestChatTransport_EditMessage_EmptyBody(t *testing.T) {
	mockSvc := new(mockChatService)
	srv := setupChatServer(t, mockSvc)
	defer srv.Close()

	bodyReader := strings.NewReader(`{"text": ""}`)
	req, err := http.NewRequest(http.MethodPut, srv.URL+"/api/v2/chats/5/messages/1", bodyReader)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	mockSvc.AssertNotCalled(t, "EditMessage")
}

func TestChatTransport_DeleteMessage_Success(t *testing.T) {
	mockSvc := new(mockChatService)
	srv := setupChatServer(t, mockSvc)
	defer srv.Close()

	mockSvc.On("SoftDeleteMessage", mock.Anything, int64(1)).Return(nil)

	req, err := http.NewRequest(http.MethodDelete, srv.URL+"/api/v2/chats/5/messages/1", nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	mockSvc.AssertExpectations(t)
}

func TestChatTransport_DeleteMessage_InvalidID(t *testing.T) {
	mockSvc := new(mockChatService)
	srv := setupChatServer(t, mockSvc)
	defer srv.Close()

	req, err := http.NewRequest(http.MethodDelete, srv.URL+"/api/v2/chats/5/messages/abc", nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	mockSvc.AssertNotCalled(t, "SoftDeleteMessage")
}

func TestChatTransport_MarkMessagesAsRead_Success(t *testing.T) {
	mockSvc := new(mockChatService)
	srv := setupChatServer(t, mockSvc)
	defer srv.Close()

	mockSvc.On("MarkMessagesAsRead", mock.Anything, int64(10), int64(42), int64(100)).Return(nil)

	bodyReader := strings.NewReader(`{"last_read_message_id": 100}`)
	req, err := http.NewRequest(http.MethodPatch, srv.URL+"/api/v2/chats/10/messages/read-status?user_id=42", bodyReader)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	mockSvc.AssertExpectations(t)
}

func TestChatTransport_MarkMessagesAsRead_MissingUserID(t *testing.T) {
	mockSvc := new(mockChatService)
	srv := setupChatServer(t, mockSvc)
	defer srv.Close()

	bodyReader := strings.NewReader(`{"last_read_message_id": 100}`)
	req, err := http.NewRequest(http.MethodPatch, srv.URL+"/api/v2/chats/10/messages/read-status", bodyReader)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	mockSvc.AssertNotCalled(t, "MarkMessagesAsRead")
}

func TestChatTransport_Pagination_InvalidLimit(t *testing.T) {
	mockSvc := new(mockChatService)
	srv := setupChatServer(t, mockSvc)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v2/chats/1/messages?limit=-1")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestChatTransport_Pagination_InvalidLastMessageID(t *testing.T) {
	mockSvc := new(mockChatService)
	srv := setupChatServer(t, mockSvc)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v2/chats/1/messages?last_message_id=abc")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestChatTransport_GetUserChatsExtended_Success(t *testing.T) {
	mockSvc := new(mockChatService)
	srv := setupChatServer(t, mockSvc)
	defer srv.Close()

	now := time.Now()
	chats := []domain.ChatInfo{
		{
			Id:   1,
			Name: "chat-1",
			LastMessage: &domain.MessageInfo{
				Id: 10, Text: "hi", SenderId: 2, SentAt: now, IsRead: false,
			},
			UnreadCount: 3,
		},
	}
	mockSvc.On("GetUserChatsExtended", mock.Anything, int64(42)).Return(chats, nil)

	resp, err := http.Get(srv.URL + "/api/v2/chats?user_id=42")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal(body, &result))
	chatsArr := result["chats"].([]any)
	require.Len(t, chatsArr, 1)
	chat := chatsArr[0].(map[string]any)
	require.Equal(t, float64(1), chat["id"])
	require.Equal(t, "chat-1", chat["name"])
	require.Equal(t, float64(3), chat["unread_count"])

	lm := chat["last_message"].(map[string]any)
	require.Equal(t, float64(10), lm["id"])
	require.Equal(t, "hi", lm["text"])
	mockSvc.AssertExpectations(t)
}
