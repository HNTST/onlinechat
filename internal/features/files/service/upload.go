package files_service

import (
	"context"
	"discord/internal/core/domain"
	"fmt"
	"io"

	"github.com/google/uuid"
)

func (s *FileService) UploadAvatar(ctx context.Context, userID int64, file io.Reader, size int64, contentType string, ext string) (int64, error) {
	objectName := fmt.Sprintf("avatars/user_%d_%s%s", userID, uuid.NewString(), ext)

	key, err := s.fileStorage.Upload(ctx, objectName, contentType, file, size)
	if err != nil {
		return 0, fmt.Errorf("upload to storage: %w", err)
	}

	var savedFile domain.Files

	err = s.fileRepository.Transaction(ctx, func(txCtx context.Context) error {
		fileDomain := domain.Files{
			FileKey:          key,
			OriginalFileName: objectName,
			TypeFile:         contentType,
			SizeFile:         size,
		}
		var innerErr error
		savedFile, innerErr = s.fileRepository.SaveFile(txCtx, fileDomain)
		if innerErr != nil {
			return fmt.Errorf("save file: %w", innerErr)
		}

		if innerErr = s.fileRepository.UpdateUserAvatar(txCtx, userID, savedFile.Id); innerErr != nil {
			return fmt.Errorf("update user avatar: %w", innerErr)
		}

		return nil
	})

	if err != nil {
		if rmErr := s.fileStorage.Remove(ctx, key); rmErr != nil {
			return 0, fmt.Errorf("upload failed: %w; cleanup error: %w", err, rmErr)
		}
		return 0, fmt.Errorf("upload failed: %w", err)
	}

	return savedFile.Id, nil
}
