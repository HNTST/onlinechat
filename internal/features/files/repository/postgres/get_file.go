package files_postgres_repository

import (
	"context"
	"discord/internal/core/domain"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (r *FileRepository) GetFileByID(ctx context.Context, id int64) (*domain.Files, error) {
	q := r.pool.GetQueries(ctx)

	file, err := q.GetFileByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get file by id: %w", err)
	}

	return &domain.Files{
		Id:               file.ID,
		FileKey:          file.FileKey,
		OriginalFileName: file.OriginalFileName,
		TypeFile:         file.TypeFile,
		SizeFile:         file.SizeFile,
	}, nil
}
