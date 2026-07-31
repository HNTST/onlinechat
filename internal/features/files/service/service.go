package files_service

import (
	"context"
	"discord/internal/core/domain"
	"io"
)

type FileService struct {
	fileRepository FileRepository
	fileStorage    FileStorage
	usersService   UserService
}

type FileRepository interface {
	Transaction(ctx context.Context, fn func(txCtx context.Context) error) error
	SaveFile(ctx context.Context, file domain.Files) (domain.Files, error)
	GetFileByID(ctx context.Context, id int64) (*domain.Files, error)
}

type FileStorage interface {
	Upload(ctx context.Context, fileName, contentType string, reader io.Reader, size int64) (string, error)
	GetAvatarStream(ctx context.Context, objectName string) (io.ReadCloser, error)
	Remove(ctx context.Context, objectName string) error
}

// UserService — интерфейс для user/service, чтобы FileService
// вызывал сервисный слой, а не репозиторий напрямую.
type UserService interface {
	UpdateAvatar(ctx context.Context, userID, fileID int64) error
}

func NewFileService(fileRepository FileRepository, fileStorage FileStorage, usersService UserService) *FileService {
	return &FileService{
		fileRepository: fileRepository,
		fileStorage:    fileStorage,
		usersService:   usersService,
	}
}
