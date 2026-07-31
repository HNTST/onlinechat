package files_service_test

import (
	"bytes"
	"context"
	"discord/internal/core/domain"
	files_service "discord/internal/features/files/service"
	"errors"
	"io"
	"regexp"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockFileRepository struct {
	mock.Mock
}

func (m *mockFileRepository) Transaction(ctx context.Context, fn func(txCtx context.Context) error) error {
	args := m.Called(ctx, fn)
	return args.Error(0)
}

func (m *mockFileRepository) SaveFile(ctx context.Context, file domain.Files) (domain.Files, error) {
	args := m.Called(ctx, file)
	return args.Get(0).(domain.Files), args.Error(1)
}

func (m *mockFileRepository) GetFileByID(ctx context.Context, id int64) (*domain.Files, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.Files), args.Error(1)
}

type mockUserService struct {
	mock.Mock
}

func (m *mockUserService) UpdateAvatar(ctx context.Context, userID, fileID int64) error {
	args := m.Called(ctx, userID, fileID)
	return args.Error(0)
}

type mockFileStorage struct {
	mock.Mock
}

func (m *mockFileStorage) Upload(ctx context.Context, fileName, contentType string, reader io.Reader, size int64) (string, error) {
	args := m.Called(ctx, fileName, contentType, reader, size)
	return args.String(0), args.Error(1)
}

func (m *mockFileStorage) GetAvatarStream(ctx context.Context, objectName string) (io.ReadCloser, error) {
	args := m.Called(ctx, objectName)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *mockFileStorage) Remove(ctx context.Context, objectName string) error {
	args := m.Called(ctx, objectName)
	return args.Error(0)
}

func TestUploadAvatar_StorageCleanupOnDatabaseError(t *testing.T) {
	mockRepo := new(mockFileRepository)
	mockStorage := new(mockFileStorage)
	mockUsers := new(mockUserService)
	svc := files_service.NewFileService(mockRepo, mockStorage, mockUsers)

	ctx := context.Background()
	fileContent := []byte("fake-image-data")
	reader := bytes.NewReader(fileContent)
	const userID int64 = 42
	size := int64(len(fileContent))
	const contentType = "image/png"
	const ext = ".png"
	expectedKey := "avatars/user_42_test-uuid.png"

	mockStorage.On("Upload", mock.Anything, mock.Anything, contentType, reader, size).
		Return(expectedKey, nil).Once()

	mockRepo.On("Transaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(errors.New("db transaction error")).Once()

	mockStorage.On("Remove", mock.Anything, expectedKey).
		Return(nil).Once()

	_, err := svc.UploadAvatar(ctx, userID, reader, size, contentType, ext)
	require.Error(t, err)

	mockStorage.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	mockStorage.AssertCalled(t, "Remove", mock.Anything, expectedKey)
}

func TestUploadAvatar_UniqueUUIDName(t *testing.T) {
	mockRepo := new(mockFileRepository)
	mockStorage := new(mockFileStorage)
	mockUsers := new(mockUserService)
	svc := files_service.NewFileService(mockRepo, mockStorage, mockUsers)

	ctx := context.Background()
	fileContent := []byte("fake-image-data")
	reader := bytes.NewReader(fileContent)
	const userID int64 = 42
	size := int64(len(fileContent))
	const contentType = "image/png"
	const ext = ".png"

	var capturedName string

	mockStorage.On("Upload", mock.Anything, mock.AnythingOfType("string"), contentType, reader, size).
		Run(func(args mock.Arguments) {
			capturedName = args.String(1)
		}).
		Return(capturedName, nil).Once()

	mockRepo.On("Transaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(args.Get(0).(context.Context))
		}).
		Return(nil).Once()

	mockRepo.On("SaveFile", mock.Anything, mock.Anything).
		Return(domain.Files{Id: 1}, nil).Once()

	mockUsers.On("UpdateAvatar", mock.Anything, userID, int64(1)).
		Return(nil).Once()

	fileID, err := svc.UploadAvatar(ctx, userID, reader, size, contentType, ext)
	require.NoError(t, err)
	require.Equal(t, int64(1), fileID)

	pattern := `^avatars/user_42_[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}\.png$`
	matched, _ := regexp.MatchString(pattern, capturedName)
	require.True(t, matched, "ObjectName должен содержать user_42 и UUID: %s", capturedName)

	mockStorage.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	mockUsers.AssertExpectations(t)
}
