package chats_postgres_repository_test

import (
	"context"
	core_errors "discord/internal/core/errors"
)

func (s *ChatRepositoryTestSuite) TestEditMessage_Success() {
	userID := s.createUser("editor", "pass")

	var chatID int64
	err := s.pool.QueryRow(context.Background(),
		"INSERT INTO chats (name) VALUES('edit-chat') RETURNING id").Scan(&chatID)
	s.Require().NoError(err)

	var msgID int64
	err = s.pool.QueryRow(context.Background(),
		"INSERT INTO messages (chat_id, sender_id, message) VALUES($1,$2,'original') RETURNING id",
		chatID, userID).Scan(&msgID)
	s.Require().NoError(err)

	err = s.repo.EditMessage(context.Background(), chatID, msgID, userID, "updated text")
	s.Require().NoError(err)

	var messageText string
	var isEdited bool
	err = s.pool.QueryRow(context.Background(),
		"SELECT message, is_edited FROM messages WHERE id=$1", msgID).Scan(&messageText, &isEdited)
	s.Require().NoError(err)
	s.Require().Equal("updated text", messageText)
	s.Require().True(isEdited)
}

func (s *ChatRepositoryTestSuite) TestEditMessage_NotFound() {
	err := s.repo.EditMessage(context.Background(), 0, 99999, 0, "text")
	s.Require().Error(err)
	s.Require().ErrorIs(err, core_errors.ErrMessageNotFound)
}

func (s *ChatRepositoryTestSuite) TestEditMessage_DeletedMessage() {
	userID := s.createUser("editor2", "pass")

	var chatID int64
	err := s.pool.QueryRow(context.Background(),
		"INSERT INTO chats (name) VALUES('edit-del-chat') RETURNING id").Scan(&chatID)
	s.Require().NoError(err)

	var msgID int64
	err = s.pool.QueryRow(context.Background(),
		"INSERT INTO messages (chat_id, sender_id, message, is_deleted) VALUES($1,$2,'deleted',true) RETURNING id",
		chatID, userID).Scan(&msgID)
	s.Require().NoError(err)

	err = s.repo.EditMessage(context.Background(), chatID, msgID, userID, "should fail")
	s.Require().Error(err)
	s.Require().ErrorIs(err, core_errors.ErrMessageNotFound)
}

func (s *ChatRepositoryTestSuite) TestEditMessage_NotOwner() {
	ownerID := s.createUser("owner", "pass")
	otherID := s.createUser("other", "pass")

	var chatID int64
	err := s.pool.QueryRow(context.Background(),
		"INSERT INTO chats (name) VALUES('edit-owner-chat') RETURNING id").Scan(&chatID)
	s.Require().NoError(err)

	var msgID int64
	err = s.pool.QueryRow(context.Background(),
		"INSERT INTO messages (chat_id, sender_id, message) VALUES($1,$2,'by-owner') RETURNING id",
		chatID, ownerID).Scan(&msgID)
	s.Require().NoError(err)

	err = s.repo.EditMessage(context.Background(), chatID, msgID, otherID, "hijack")
	s.Require().Error(err)
	s.Require().ErrorIs(err, core_errors.ErrUnauthorized)
}

func (s *ChatRepositoryTestSuite) TestEditMessage_WrongChat() {
	userID := s.createUser("editor3", "pass")

	var chatID int64
	err := s.pool.QueryRow(context.Background(),
		"INSERT INTO chats (name) VALUES('edit-chat3') RETURNING id").Scan(&chatID)
	s.Require().NoError(err)

	var msgID int64
	err = s.pool.QueryRow(context.Background(),
		"INSERT INTO messages (chat_id, sender_id, message) VALUES($1,$2,'own-msg') RETURNING id",
		chatID, userID).Scan(&msgID)
	s.Require().NoError(err)

	err = s.repo.EditMessage(context.Background(), chatID+999, msgID, userID, "wrong chat")
	s.Require().Error(err)
	s.Require().ErrorIs(err, core_errors.ErrUnauthorized)
}
