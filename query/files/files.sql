-- name: InsertFile :one
INSERT INTO files (file_key, original_file_name, type_file, size_file)
VALUES ($1, $2, $3, $4)
RETURNING id, file_key, original_file_name, type_file, size_file;

-- name: UpdateUserAvatarByID :exec
UPDATE users SET avatar_id = $1 WHERE id = $2;

-- name: GetFileByID :one
SELECT id, file_key, original_file_name, type_file, size_file
FROM files
WHERE id = $1;
