package files_service

import (
	"context"
	"fmt"
	"io"
)

func (s *FileService) GetAvatarStream(ctx context.Context, objectName string) (io.ReadCloser, error) {
	stream, err := s.fileStorage.GetAvatarStream(ctx, objectName)
	if err != nil {
		return nil, fmt.Errorf("get avatar from storage: %w", err)
	}
	return stream, nil
}

func (s *FileService) GetAvatarByID(ctx context.Context, id int64) (io.ReadCloser, string, error) {
	file, err := s.fileRepository.GetFileByID(ctx, id)
	if err != nil {
		return nil, "", fmt.Errorf("get file by id: %w", err)
	}
	if file == nil {
		return nil, "", fmt.Errorf("file not found")
	}

	stream, err := s.fileStorage.GetAvatarStream(ctx, file.FileKey)
	if err != nil {
		return nil, "", fmt.Errorf("get avatar from storage: %w", err)
	}

	return stream, file.TypeFile, nil
}
