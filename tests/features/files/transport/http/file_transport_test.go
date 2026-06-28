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
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockFileService struct {
	mock.Mock
}

func (m *mockFileService) UploadAvatar(ctx context.Context, userID int64, file io.Reader, size int64, contentType string, ext string) (int64, error) {
	args := m.Called(ctx, userID, file, size, contentType, ext)
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

func uploadTestFile(t *testing.T, srv *httptest.Server, mockSvc *mockFileService, fileContent []byte, fileName string) *http.Response {
	t.Helper()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	part, err := w.CreateFormFile("new_user_avatar", fileName)
	require.NoError(t, err)

	_, err = part.Write(fileContent)
	require.NoError(t, err)
	require.NoError(t, w.Close())

	req, err := http.NewRequest(http.MethodPost, srv.URL+"/api/v2/user/avatar?user_id=1", &buf)
	require.NoError(t, err)
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func TestUploadAvatar_SizeLimitExceeded(t *testing.T) {
	mockSvc := new(mockFileService)
	srv := setupFileServer(t, mockSvc)
	defer srv.Close()

	fileContent := make([]byte, 6*1024*1024)
	for i := range fileContent {
		fileContent[i] = 0xFF
	}

	resp := uploadTestFile(t, srv, mockSvc, fileContent, "photo.jpg")
	defer resp.Body.Close()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]string
	require.NoError(t, json.Unmarshal(body, &result))
	require.Contains(t, result["error"], "неверный аргумент")
	mockSvc.AssertNotCalled(t, "UploadAvatar")
}

func TestUploadAvatar_InvalidMIMEType(t *testing.T) {
	mockSvc := new(mockFileService)
	srv := setupFileServer(t, mockSvc)
	defer srv.Close()

	fileContent := []byte("MZ\x90\x00\x03\x00\x00\x00\x04\x00\x00\x00\xff\xff\x00\x00")

	resp := uploadTestFile(t, srv, mockSvc, fileContent, "image.exe")
	defer resp.Body.Close()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]string
	require.NoError(t, json.Unmarshal(body, &result))
	require.Contains(t, result["error"], "неверный аргумент")
	mockSvc.AssertNotCalled(t, "UploadAvatar")
}

func TestUploadAvatar_Success(t *testing.T) {
	mockSvc := new(mockFileService)
	srv := setupFileServer(t, mockSvc)
	defer srv.Close()

	fileContent := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
	}

	mockSvc.On("UploadAvatar", mock.Anything, int64(1), mock.Anything,
		int64(len(fileContent)), "image/png", ".png").
		Return(int64(7), nil).Once()

	resp := uploadTestFile(t, srv, mockSvc, fileContent, "avatar.png")
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal(body, &result))
	require.Equal(t, float64(7), result["file_id"])
	mockSvc.AssertExpectations(t)
}
