package chats_postgres_repository

import (
	"context"
	"discord/internal/core/domain"
	"database/sql"
	"fmt"
)

func (r *ChatRepository) GetUserChatsExtended(ctx context.Context, userID int64) ([]domain.ChatInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT
			c.id,
			c.name,
			lm.id,
			lm.message,
			lm.sender_id,
			lm.sent_at,
			lm.is_read,
			COALESCE(uc.unread_count, 0)
		FROM chats c
		JOIN chat_members cm ON c.id = cm.chat_id
		LEFT JOIN LATERAL (
			SELECT id, message, sender_id, sent_at, is_read
			FROM messages
			WHERE chat_id = c.id AND is_deleted = false
			ORDER BY id DESC
			LIMIT 1
		) lm ON true
		LEFT JOIN LATERAL (
			SELECT COUNT(*) AS unread_count
			FROM messages
			WHERE chat_id = c.id AND sender_id != $2 AND is_read = false
		) uc ON true
		WHERE cm.user_id = $1
		ORDER BY c.id
	`
	rows, err := r.pool.Query(ctx, query, userID, userID)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	var result []domain.ChatInfo
	for rows.Next() {
		var (
			ci           domain.ChatInfo
			lmID         sql.NullInt64
			lmText       sql.NullString
			lmSenderID   sql.NullInt64
			lmSentAt     sql.NullTime
			lmIsRead     sql.NullBool
		)
		if err := rows.Scan(
			&ci.Id, &ci.Name,
			&lmID, &lmText, &lmSenderID, &lmSentAt, &lmIsRead,
			&ci.UnreadCount,
		); err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		if lmID.Valid {
			ci.LastMessage = &domain.MessageInfo{
				Id:       lmID.Int64,
				Text:     lmText.String,
				SenderId: lmSenderID.Int64,
				SentAt:   lmSentAt.Time,
				IsRead:   lmIsRead.Bool,
			}
		}
		result = append(result, ci)
	}
	return result, nil
}
