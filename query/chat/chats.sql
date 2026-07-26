-- name: InsertChat :one
INSERT INTO chats (name, is_group)
VALUES ($1, $2)
RETURNING id;

-- name: DeleteChatByID :exec
DELETE FROM chats WHERE id = $1;

-- name: UpdateChatNameByID :exec
UPDATE chats SET name = $1 WHERE id = $2;
