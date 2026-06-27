package files_service

import (
	"context"
	"discord/internal/core/domain"
	"io"
)

type FileService struct {
	fileRepository FileRepository
	fileStorage    FileStorage
}

type FileRepository interface {
	SaveFile(ctx context.Context, file domain.Files) (domain.Files, error)
	UpdateUserAvatar(ctx context.Context, userID, fileID int64) error
	GetFileByID(ctx context.Context, id int64) (*domain.Files, error)
}

type FileStorage interface {
	Upload(ctx context.Context, fileName, contentType string, reader io.Reader, size int64) (string, error)
	GetAvatarStream(ctx context.Context, objectName string) (io.ReadCloser, error)
}

func NewFileService(fileRepository FileRepository, fileStorage FileStorage) *FileService {
	return &FileService{
		fileRepository: fileRepository,
		fileStorage:    fileStorage,
	}
}
