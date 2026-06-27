package files_repository_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	config "discord/internal/core/config"
	files_repository "discord/internal/features/files/repository"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
)

type MinioRepositoryTestSuite struct {
	suite.Suite
	storage *files_repository.MinioStorage
	ctx     context.Context
}

func TestMinioRepositorySuite(t *testing.T) {
	suite.Run(t, new(MinioRepositoryTestSuite))
}

func (s *MinioRepositoryTestSuite) SetupSuite() {
	_ = godotenv.Load("../../../../internal/core/config/.env.test")

	s.ctx = context.Background()

	cfg := config.Config_Minio{
		Host:         getEnv("MINIO_HOST", "localhost:9000"),
		BucketName:   getEnv("MINIO_BUCKET_NAME", "discord-avatars-test"),
		RootUser:     getEnv("MINIO_ROOT_USER", "test_admin"),
		RootPassword: getEnv("MINIO_ROOT_PASSWORD", "test_password"),
	}

	storage, err := files_repository.NewMinioStorage(cfg)
	s.Require().NoError(err, "NewMinioStorage должен создать клиент и бакет")
	s.Require().NotNil(storage)
	s.storage = storage
}

func (s *MinioRepositoryTestSuite) TearDownSuite() {
	if s.storage != nil {
		s.cleanupBucket()
	}
}

func (s *MinioRepositoryTestSuite) cleanupBucket() {
	// not implemented via public API; bucket will be removed by docker compose down -v
}

func (s *MinioRepositoryTestSuite) TestUploadAndGetAvatarStream_Success() {
	// Arrange
	fileKey := "avatars/test-avatar.png"
	contentType := "image/png"
	content := []byte("fake-png-binary-content-12345")
	reader := bytes.NewReader(content)

	// Act — Upload
	key, err := s.storage.Upload(s.ctx, fileKey, contentType, reader, int64(len(content)))
	s.Require().NoError(err, "Upload должен сохранить файл")
	s.Require().Equal(fileKey, key)

	// Act — GetAvatarStream
	stream, err := s.storage.GetAvatarStream(s.ctx, fileKey)
	s.Require().NoError(err, "GetAvatarStream должен найти файл")
	defer stream.Close()

	downloaded, err := io.ReadAll(stream)
	s.Require().NoError(err, "Чтение потока не должно вызывать ошибку")

	// Assert — байты совпадают
	s.Require().Equal(content, downloaded, "Скачанные байты должны совпадать с исходными")
}

func (s *MinioRepositoryTestSuite) TestGetAvatarStream_NotFound() {
	_, err := s.storage.GetAvatarStream(s.ctx, "avatars/nonexistent-file.png")
	s.Require().Error(err, "Должна быть ошибка для несуществующего файла")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
