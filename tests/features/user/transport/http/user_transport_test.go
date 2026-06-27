package users_transport_http_test

import (
	"context"
	"discord/internal/core/domain"
	core_errors "discord/internal/core/errors"
	core_logger "discord/internal/core/logger"
	core_http_middleware "discord/internal/core/transport/middleware"
	core_http_server "discord/internal/core/transport/server"
	users_transport_http "discord/internal/features/user/transport/http"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockUsersService struct {
	mock.Mock
}

func (m *mockUsersService) SaveUser(ctx context.Context, user domain.User) (domain.User, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(domain.User), args.Error(1)
}

func (m *mockUsersService) GetUser(ctx context.Context, user_id *int) (domain.User, error) {
	args := m.Called(ctx, user_id)
	return args.Get(0).(domain.User), args.Error(1)
}

func (m *mockUsersService) GetUserID(ctx context.Context, username string) (int64, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(int64), args.Error(1)
}

func setupUserServer(t *testing.T, svc *mockUsersService) *httptest.Server {
	t.Helper()
	logger := core_logger.NewLogger()
	handler := users_transport_http.NewUsersHTTPHandler(svc)
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

func TestUserTransport_GetUser_Success(t *testing.T) {
	mockSvc := new(mockUsersService)
	srv := setupUserServer(t, mockSvc)
	defer srv.Close()

	avatarID := int64(42)
	expectedUser := domain.User{
		Id:       1,
		Name:     "ivan",
		Password: "-1",
		Reg:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		AvatarID: &avatarID,
	}

	mockSvc.On("GetUser", mock.Anything, mock.MatchedBy(func(p *int) bool {
		return p != nil && *p == 1
	})).Return(expectedUser, nil)

	resp, err := http.Get(srv.URL + "/api/v2/user?user_id=1")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal(body, &result))

	require.Equal(t, float64(1), result["id"])
	require.Equal(t, "ivan", result["username"])
	require.Equal(t, float64(42), result["avatar_id"])
	mockSvc.AssertExpectations(t)
}

func TestUserTransport_GetUser_NotFound(t *testing.T) {
	mockSvc := new(mockUsersService)
	srv := setupUserServer(t, mockSvc)
	defer srv.Close()

	mockSvc.On("GetUser", mock.Anything, mock.MatchedBy(func(p *int) bool {
		return p != nil && *p == 999
	})).Return(domain.User{}, fmt.Errorf("user not found: %w", core_errors.ErrUserNotFound))

	resp, err := http.Get(srv.URL + "/api/v2/user?user_id=999")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusNotFound, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]string
	require.NoError(t, json.Unmarshal(body, &result))
	require.Contains(t, result["error"], "не найден")
	mockSvc.AssertExpectations(t)
}

func TestUserTransport_GetUser_ErrorMasking(t *testing.T) {
	mockSvc := new(mockUsersService)
	srv := setupUserServer(t, mockSvc)
	defer srv.Close()

	rawError := errors.New("connection refused — db unavailable")
	mockSvc.On("GetUser", mock.Anything, mock.MatchedBy(func(p *int) bool {
		return p != nil && *p == 1
	})).Return(domain.User{}, rawError)

	resp, err := http.Get(srv.URL + "/api/v2/user?user_id=1")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]string
	require.NoError(t, json.Unmarshal(body, &result))
	require.Equal(t, "Внутренняя ошибка сервера", result["error"])
	require.NotContains(t, string(body), "connection refused")
	require.NotContains(t, string(body), "db unavailable")
	mockSvc.AssertExpectations(t)
}
