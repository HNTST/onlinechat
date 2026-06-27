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
