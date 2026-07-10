package auth_transport_http_test

import (
	"bytes"
	"context"
	"discord/internal/core/domain"
	core_logger "discord/internal/core/logger"
	core_http_middleware "discord/internal/core/transport/middleware"
	core_http_server "discord/internal/core/transport/server"
	auth_transport_http "discord/internal/features/auth/transport/http"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockAuthService struct {
	mock.Mock
}

func (m *mockAuthService) Login(ctx context.Context, username, password string) (string, domain.User, error) {
	args := m.Called(ctx, username, password)
	return args.String(0), args.Get(1).(domain.User), args.Error(2)
}

func (m *mockAuthService) Logout(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *mockAuthService) ValidateSession(ctx context.Context, token string) (int64, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(int64), args.Error(1)
}

func setupAuthServer(t *testing.T, svc *mockAuthService) *httptest.Server {
	t.Helper()
	logger := core_logger.NewLogger()
	handler := auth_transport_http.NewAuthHTTPHandler(svc)
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

func TestAuthTransport_Login_Success(t *testing.T) {
	mockSvc := new(mockAuthService)
	srv := setupAuthServer(t, mockSvc)
	defer srv.Close()

	expectedUser := domain.User{
		Id:   1,
		Name: "alice",
		Reg:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	mockSvc.On("Login", mock.Anything, "alice", "secret123").
		Return("session-token-abc", expectedUser, nil)

	body := `{"username":"alice","password":"secret123"}`
	resp, err := http.Post(srv.URL+"/api/v2/auth/login", "application/json", bytes.NewBufferString(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal(respBody, &result))
	require.Equal(t, float64(1), result["id"])
	require.Equal(t, "alice", result["username"])

	cookies := resp.Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "session_token" {
			sessionCookie = c
			break
		}
	}
	require.NotNil(t, sessionCookie)
	require.Equal(t, "session-token-abc", sessionCookie.Value)
	require.True(t, sessionCookie.HttpOnly)
	require.Equal(t, "/", sessionCookie.Path)

	mockSvc.AssertExpectations(t)
}

func TestAuthTransport_Login_InvalidCredentials(t *testing.T) {
	mockSvc := new(mockAuthService)
	srv := setupAuthServer(t, mockSvc)
	defer srv.Close()

	mockSvc.On("Login", mock.Anything, "alice", "wrongpass").
		Return("", domain.User{}, errors.New("не авторизован"))

	body := `{"username":"alice","password":"wrongpass"}`
	resp, err := http.Post(srv.URL+"/api/v2/auth/login", "application/json", bytes.NewBufferString(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	mockSvc.AssertExpectations(t)
}

func TestAuthTransport_Login_BadRequest(t *testing.T) {
	mockSvc := new(mockAuthService)
	srv := setupAuthServer(t, mockSvc)
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/api/v2/auth/login", "application/json", bytes.NewBufferString(`not json`))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	mockSvc.AssertNotCalled(t, "Login")
}

func TestAuthTransport_Logout_Success(t *testing.T) {
	mockSvc := new(mockAuthService)
	srv := setupAuthServer(t, mockSvc)
	defer srv.Close()

	mockSvc.On("Logout", mock.Anything, "test-token").Return(nil)

	req, err := http.NewRequest(http.MethodPost, srv.URL+"/api/v2/auth/logout", nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: "test-token"})

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	cookies := resp.Cookies()
	var clearedCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "session_token" {
			clearedCookie = c
			break
		}
	}
	require.NotNil(t, clearedCookie)
	require.Equal(t, "", clearedCookie.Value)
	require.Equal(t, -1, clearedCookie.MaxAge)

	mockSvc.AssertExpectations(t)
}

func TestAuthTransport_Logout_NoCookie(t *testing.T) {
	mockSvc := new(mockAuthService)
	srv := setupAuthServer(t, mockSvc)
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/api/v2/auth/logout", "application/json", nil)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	mockSvc.AssertNotCalled(t, "Logout")
}
