package chats_postgres_repository

import (
	"context"
	"discord/internal/core/domain"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v5"
)

func (r *ChatRepository) GetChatMessages(ctx context.Context, chatID int64) ([]domain.Messages, error) {
	q := r.pool.GetQueries(ctx)
	rows, err := q.GetMessagesByChatID(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	result := make([]domain.Messages, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.Messages{
			Id:        row.ID,
			SenderId:  row.SenderID,
			Message:   row.Message,
			Sent_at:   row.SentAt,
			IsEdited:  row.IsEdited,
			IsDeleted: row.IsDeleted,
			IsRead:    row.IsRead,
		})
	}
	return result, nil
}

func (r *ChatRepository) GetChatMessagesPaginated(
	ctx context.Context,
	chatID int64,
	lastMessageID *int64,
	limit int,
) ([]domain.Messages, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
		select id, sender_id, message, sent_at, is_edited, is_deleted, is_read
		from messages
		where chat_id = $1
	`
	var args []any
	args = append(args, chatID)

	if lastMessageID != nil {
		query += ` AND id < $2`
		args = append(args, *lastMessageID)
	}
	query += ` ORDER BY id DESC LIMIT $` + fmt.Sprintf("%d", len(args)+1)
	args = append(args, limit)

	var (
		rws pgx.Rows
		err error
	)

	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	rws, err = r.pool.Query(ctx, query, args...)

	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rws.Close()

	var result []domain.Messages
	for rws.Next() {
		var m domain.Messages
		if err := rws.Scan(&m.Id, &m.SenderId, &m.Message, &m.Sent_at, &m.IsEdited, &m.IsDeleted, &m.IsRead); err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		result = append(result, m)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Id < result[j].Id
	})

	return result, nil
}
