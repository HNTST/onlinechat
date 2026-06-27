package chats_postgres_repository_test

import (
	"context"

	"discord/internal/core/domain"
)

func (s *ChatRepositoryTestSuite) TestSaveMessage_Success() {
	userID := s.createUser("sender", "pass")

	var chatID int64
	err := s.pool.QueryRow(context.Background(),
		"INSERT INTO chats (name) VALUES('save-msg-chat') RETURNING id").Scan(&chatID)
	s.Require().NoError(err)

	msg := domain.Messages{
		ChatId:   chatID,
		SenderId: userID,
		Message:  "hello world",
	}
	err = s.repo.SaveMessage(context.Background(), msg)

	s.Require().NoError(err)

	var count int
	err = s.pool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM messages WHERE chat_id=$1", chatID).Scan(&count)
	s.Require().NoError(err)
	s.Require().Equal(1, count)
}


