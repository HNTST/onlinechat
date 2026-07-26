package chats_postgres_repository_test

import (
	"context"
	core_errors "discord/internal/core/errors"
)

func (s *ChatRepositoryTestSuite) TestSoftDeleteMessage_Success() {
	userID := s.createUser("deleter", "pass")

	var chatID int64
	err := s.pool.QueryRow(context.Background(),
		"INSERT INTO chats (name) VALUES('del-chat') RETURNING id").Scan(&chatID)
	s.Require().NoError(err)

	var msgID int64
	err = s.pool.QueryRow(context.Background(),
		"INSERT INTO messages (chat_id, sender_id, message) VALUES($1,$2,'to-delete') RETURNING id",
		chatID, userID).Scan(&msgID)
	s.Require().NoError(err)

	err = s.repo.SoftDeleteMessage(context.Background(), chatID, msgID)
	s.Require().NoError(err)

	var messageText string
	var isDeleted bool
	err = s.pool.QueryRow(context.Background(),
		"SELECT message, is_deleted FROM messages WHERE id=$1", msgID).Scan(&messageText, &isDeleted)
	s.Require().NoError(err)
	s.Require().True(isDeleted)
	s.Require().Empty(messageText)
}

func (s *ChatRepositoryTestSuite) TestSoftDeleteMessage_NotFound() {
	err := s.repo.SoftDeleteMessage(context.Background(), 99998, 99999)
	s.Require().Error(err)
	s.Require().ErrorIs(err, core_errors.ErrMessageNotFound)
}

func (s *ChatRepositoryTestSuite) TestSoftDeleteMessage_WrongChat() {
	userID := s.createUser("deleter3", "pass")

	var chatID int64
	err := s.pool.QueryRow(context.Background(),
		"INSERT INTO chats (name) VALUES('del-chat3') RETURNING id").Scan(&chatID)
	s.Require().NoError(err)

	var msgID int64
	err = s.pool.QueryRow(context.Background(),
		"INSERT INTO messages (chat_id, sender_id, message) VALUES($1,$2,'wrong-chat-msg') RETURNING id",
		chatID, userID).Scan(&msgID)
	s.Require().NoError(err)

	err = s.repo.SoftDeleteMessage(context.Background(), chatID+999, msgID)
	s.Require().Error(err)
	s.Require().ErrorIs(err, core_errors.ErrUnauthorized)
}

func (s *ChatRepositoryTestSuite) TestSoftDeleteMessage_AlreadyDeleted() {
	userID := s.createUser("deleter2", "pass")

	var chatID int64
	err := s.pool.QueryRow(context.Background(),
		"INSERT INTO chats (name) VALUES('del-chat2') RETURNING id").Scan(&chatID)
	s.Require().NoError(err)

	var msgID int64
	err = s.pool.QueryRow(context.Background(),
		"INSERT INTO messages (chat_id, sender_id, message, is_deleted) VALUES($1,$2,'already-del',true) RETURNING id",
		chatID, userID).Scan(&msgID)
	s.Require().NoError(err)

	err = s.repo.SoftDeleteMessage(context.Background(), chatID, msgID)
	s.Require().Error(err)
	s.Require().ErrorIs(err, core_errors.ErrMessageNotFound)
}
