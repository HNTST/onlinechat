package chats_postgres_repository_test

import (
	"context"
)

func (s *ChatRepositoryTestSuite) TestGetChatMessages_Success() {
	userID := s.createUser("msg-sender", "pass")

	var chatID int64
	err := s.pool.QueryRow(context.Background(),
		"INSERT INTO chats (name) VALUES('msg-chat') RETURNING id").Scan(&chatID)
	s.Require().NoError(err)

	_, err = s.pool.Exec(context.Background(),
		"INSERT INTO messages (chat_id, sender_id, message) VALUES($1,$2,'hi'),($1,$2,'there')",
		chatID, userID)
	s.Require().NoError(err)

	msgs, err := s.repo.GetChatMessages(context.Background(), chatID)

	s.Require().NoError(err)
	s.Require().Len(msgs, 2)
	s.Require().Equal("hi", msgs[0].Message)
	s.Require().Equal("there", msgs[1].Message)
}

func (s *ChatRepositoryTestSuite) TestGetChatMessages_NoMessages() {
	var chatID int64
	err := s.pool.QueryRow(context.Background(),
		"INSERT INTO chats (name) VALUES('empty-chat') RETURNING id").Scan(&chatID)
	s.Require().NoError(err)

	msgs, err := s.repo.GetChatMessages(context.Background(), chatID)

	s.Require().NoError(err)
	s.Require().Empty(msgs)
}
