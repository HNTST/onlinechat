-- name: InsertFile :one
INSERT INTO files (file_key, original_file_name, type_file, size_file)
VALUES ($1, $2, $3, $4)
RETURNING id, file_key, original_file_name, type_file, size_file;

-- name: GetFileByID :one
SELECT id, file_key, original_file_name, type_file, size_file
FROM files
WHERE id = $1;
