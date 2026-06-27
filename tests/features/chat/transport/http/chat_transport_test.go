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
