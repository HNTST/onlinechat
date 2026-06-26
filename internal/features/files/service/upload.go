package files_service

import (
	"context"
	"discord/internal/core/domain"
	"fmt"
	"io"
)

func (s *FileService) UploadAvatar(ctx context.Context, userID int64, file io.Reader, origFileName string, size int64, contentType string) (int64, error) {
	fileName := fmt.Sprintf("avatars/user_%d", userID)

	key, err := s.fileStorage.Upload(ctx, fileName, contentType, file, size)
	if err != nil {
		return 0, fmt.Errorf("upload to storage: %w", err)
	}

	fileDomain := domain.Files{
		FileKey:          key,
		OriginalFileName: origFileName,
		TypeFile:         contentType,
		SizeFile:         size,
	}
	savedFile, err := s.fileRepository.SaveFile(ctx, fileDomain)
	if err != nil {
		return 0, fmt.Errorf("save file: %w", err)
	}

	if err := s.fileRepository.UpdateUserAvatar(ctx, userID, savedFile.Id); err != nil {
		return 0, fmt.Errorf("update user avatar: %w", err)
	}

	return savedFile.Id, nil
}
