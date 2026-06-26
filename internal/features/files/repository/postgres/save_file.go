package files_postgres_repository

import (
	"context"
	"discord/internal/core/domain"
	"fmt"
)

func (r *FileRepository) SaveFile(ctx context.Context, file domain.Files) (domain.Files, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		INSERT INTO files (file_key, original_file_name, type_file, size_file)
		VALUES($1, $2, $3, $4)
		RETURNING id, file_key, original_file_name, type_file, size_file
	`
	row := r.pool.QueryRow(ctx, query, file.FileKey, file.OriginalFileName, file.TypeFile, file.SizeFile)
	var result domain.Files
	err := row.Scan(&result.Id, &result.FileKey, &result.OriginalFileName, &result.TypeFile, &result.SizeFile)
	if err != nil {
		return domain.Files{}, fmt.Errorf("scan error: %w", err)
	}
	return result, nil
}

func (r *FileRepository) UpdateUserAvatar(ctx context.Context, userID, fileID int64) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `UPDATE users SET avatar_id = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, fileID, userID)
	if err != nil {
		return fmt.Errorf("exec error: %w", err)
	}
	return nil
}
