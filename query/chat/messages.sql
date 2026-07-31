-- name: InsertMessage :one
INSERT INTO messages (chat_id, sender_id, message, sent_at)
VALUES ($1, $2, $3, $4)
RETURNING id;

-- name: SoftDeleteMessage :exec
UPDATE messages
SET is_deleted = true, message = ''
WHERE id = $1 AND chat_id = $2 AND is_deleted = false;

-- name: CheckMessageExists :one
SELECT EXISTS(SELECT 1 FROM messages WHERE id = $1);

-- name: CheckMessageExistsInChat :one
SELECT EXISTS(SELECT 1 FROM messages WHERE id = $1 AND chat_id = $2);

-- name: GetMessageOwner :one
SELECT sender_id, chat_id, is_deleted
FROM messages
WHERE id = $1;

-- name: EditMessageText :exec
UPDATE messages
SET message = $1, is_edited = true
WHERE id = $2 AND is_deleted = false;

-- name: GetMessagesByChatID :many
SELECT id, sender_id, message, sent_at, is_edited, is_deleted, is_read
FROM messages
WHERE chat_id = $1
ORDER BY id ASC;

-- name: MarkMessagesReadByChatID :exec
UPDATE messages
SET is_read = true
WHERE chat_id = $1 AND sender_id != $2 AND id <= $3 AND is_read = false;

-- name: GetMessagesByChatIDDesc :many
SELECT id, sender_id, message, sent_at, is_edited, is_deleted, is_read
FROM messages
WHERE chat_id = $1 AND id < $2
ORDER BY id DESC
LIMIT $3;
