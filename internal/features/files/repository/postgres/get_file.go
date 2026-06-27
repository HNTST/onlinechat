package files_postgres_repository

import (
	"context"
	"discord/internal/core/domain"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (r *FileRepository) GetFileByID(ctx context.Context, id int64) (*domain.Files, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT id, file_key, original_file_name, type_file, size_file
		FROM files
		WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, query, id)
	var file domain.Files
	err := row.Scan(&file.Id, &file.FileKey, &file.OriginalFileName, &file.TypeFile, &file.SizeFile)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan error: %w", err)
	}
	return &file, nil
}
