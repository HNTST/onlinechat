-- name: GetUserByID :one
SELECT id, username, password, reg, avatar_id
FROM users
WHERE id = $1;

-- name: GetUserIDByUsername :one
SELECT id FROM users WHERE username = $1;

-- name: InsertUser :one
INSERT INTO users (username, password, reg)
VALUES ($1, $2, $3)
RETURNING id, username, password, reg, avatar_id;

-- name: UpdateUserLastSeen :exec
UPDATE users SET last_seen_at = $1 WHERE id = $2;

-- name: GetUserLastSeenByID :one
SELECT last_seen_at FROM users WHERE id = $1;

-- name: UpdateUserAvatarByID :exec
UPDATE users SET avatar_id = $1 WHERE id = $2;
