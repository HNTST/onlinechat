-- name: InsertChatMember :exec
INSERT INTO chat_members (chat_id, user_id)
SELECT $1, unnest($2::integer[])
ON CONFLICT (chat_id, user_id) DO NOTHING;

-- name: CheckChatExistsBetweenUsers :one
SELECT chat_id
FROM chat_members
WHERE user_id IN ($1, $2)
GROUP BY chat_id
HAVING COUNT(DISTINCT user_id) = 2;

-- name: GetChatMembersByChatID :many
SELECT user_id FROM chat_members WHERE chat_id = $1;

-- name: GetDistinctChatPartnerIDs :many
SELECT DISTINCT cm2.user_id
FROM chat_members cm1
JOIN chat_members cm2 ON cm1.chat_id = cm2.chat_id
WHERE cm1.user_id = $1 AND cm2.user_id != $1;

-- name: GetChatsByUserID :many
SELECT c.id, c.name, c.is_group
FROM chats c
JOIN chat_members cm ON c.id = cm.chat_id
WHERE cm.user_id = $1;

-- name: GetChatWithMembership :one
SELECT c.id, c.name, c.is_group
FROM chats c
JOIN chat_members cm ON c.id = cm.chat_id
WHERE c.id = $1 AND cm.user_id = $2;

-- name: GetUserChatIDs :many
SELECT chat_id FROM chat_members WHERE user_id = $1;
