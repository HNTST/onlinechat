package files_postgres_repository

import (
	"context"
	"discord/internal/core/domain"
	"discord/internal/db"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

func (r *FileRepository) SaveFile(ctx context.Context, file domain.Files) (domain.Files, error) {
	q := r.pool.GetQueries(ctx)

	result, err := q.InsertFile(ctx, db.InsertFileParams{
		FileKey:          file.FileKey,
		OriginalFileName: file.OriginalFileName,
		TypeFile:         file.TypeFile,
		SizeFile:         file.SizeFile,
	})
	if err != nil {
		return domain.Files{}, fmt.Errorf("insert file: %w", err)
	}

	return domain.Files{
		Id:               result.ID,
		FileKey:          result.FileKey,
		OriginalFileName: result.OriginalFileName,
		TypeFile:         result.TypeFile,
		SizeFile:         result.SizeFile,
	}, nil
}

func (r *FileRepository) UpdateUserAvatar(ctx context.Context, userID, fileID int64) error {
	q := r.pool.GetQueries(ctx)

	err := q.UpdateUserAvatarByID(ctx, db.UpdateUserAvatarByIDParams{
		AvatarID: pgtype.Int8{Int64: fileID, Valid: true},
		ID:       userID,
	})
	if err != nil {
		return fmt.Errorf("update user avatar: %w", err)
	}
	return nil
}
