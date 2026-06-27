package files_transport_http_test

import (
	"bytes"
	"context"
	core_errors "discord/internal/core/errors"
	core_logger "discord/internal/core/logger"
	core_http_middleware "discord/internal/core/transport/middleware"
	core_http_server "discord/internal/core/transport/server"
	files_transport_http "discord/internal/features/files/transport/http"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockFileService struct {
	mock.Mock
}

func (m *mockFileService) UploadAvatar(ctx context.Context, userID int64, file io.Reader, origFileName string, size int64, contentType string) (int64, error) {
	args := m.Called(ctx, userID, file, origFileName, size, contentType)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockFileService) GetAvatarStream(ctx context.Context, objectName string) (io.ReadCloser, error) {
	args := m.Called(ctx, objectName)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *mockFileService) GetAvatarByID(ctx context.Context, id int64) (io.ReadCloser, string, error) {
	args := m.Called(ctx, id)
	var rc io.ReadCloser
	if a := args.Get(0); a != nil {
		rc = a.(io.ReadCloser)
	}
	return rc, args.String(1), args.Error(2)
}

func setupFileServer(t *testing.T, svc *mockFileService) *httptest.Server {
	t.Helper()
	logger := core_logger.NewLogger()
	handler := files_transport_http.NewFileHTTPHandler(svc)
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

func TestFileTransport_GetAvatarByID_Success(t *testing.T) {
	mockSvc := new(mockFileService)
	srv := setupFileServer(t, mockSvc)
	defer srv.Close()

	imageData := []byte("fake-png-binary-data")
	rc := io.NopCloser(bytes.NewReader(imageData))

	mockSvc.On("GetAvatarByID", mock.Anything, int64(1)).Return(rc, "image/png", nil)

	resp, err := http.Get(srv.URL + "/api/v2/user/avatar?id=1")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "image/png", resp.Header.Get("Content-Type"))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, imageData, body)
	mockSvc.AssertExpectations(t)
}

func TestFileTransport_GetAvatarByID_NotFound(t *testing.T) {
	mockSvc := new(mockFileService)
	srv := setupFileServer(t, mockSvc)
	defer srv.Close()

	mockSvc.On("GetAvatarByID", mock.Anything, int64(999)).Return(nil, "", fmt.Errorf("file not found: %w", core_errors.ErrUserNotFound))

	resp, err := http.Get(srv.URL + "/api/v2/user/avatar?id=999")
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

func TestFileTransport_GetAvatarByID_MissingParam(t *testing.T) {
	mockSvc := new(mockFileService)
	srv := setupFileServer(t, mockSvc)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v2/user/avatar")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]string
	require.NoError(t, json.Unmarshal(body, &result))
	require.Contains(t, result["error"], "неверный аргумент")
	mockSvc.AssertNotCalled(t, "GetAvatarByID")
}
