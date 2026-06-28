package chats_postgres_repository_test

import (
	"context"
)

func (s *ChatRepositoryTestSuite) TestGetUserChatsExtended_Success() {
	userA := s.createUser("ext-user-a", "pass")
	userB := s.createUser("ext-user-b", "pass")

	var chatID int64
	err := s.pool.QueryRow(context.Background(),
		"INSERT INTO chats (name) VALUES('ext-chat') RETURNING id").Scan(&chatID)
	s.Require().NoError(err)

	_, err = s.pool.Exec(context.Background(),
		"INSERT INTO chat_members (chat_id, user_id) VALUES($1,$2),($1,$3)",
		chatID, userA, userB)
	s.Require().NoError(err)

	_, err = s.pool.Exec(context.Background(),
		"INSERT INTO messages (chat_id, sender_id, message, is_read) VALUES($1,$2,'last-msg',false)",
		chatID, userB)
	s.Require().NoError(err)

	chats, err := s.repo.GetUserChatsExtended(context.Background(), userA)
	s.Require().NoError(err)
	s.Require().Len(chats, 1)
	s.Require().Equal("ext-chat", chats[0].Name)
	s.Require().NotNil(chats[0].LastMessage)
	s.Require().Equal("last-msg", chats[0].LastMessage.Text)
	s.Require().Equal(userB, chats[0].LastMessage.SenderId)
	s.Require().Equal(1, chats[0].UnreadCount)
}

func (s *ChatRepositoryTestSuite) TestGetUserChatsExtended_NoMessages() {
	userA := s.createUser("ext-empty-user", "pass")

	var chatID int64
	err := s.pool.QueryRow(context.Background(),
		"INSERT INTO chats (name) VALUES('empty-ext-chat') RETURNING id").Scan(&chatID)
	s.Require().NoError(err)

	_, err = s.pool.Exec(context.Background(),
		"INSERT INTO chat_members (chat_id, user_id) VALUES($1,$2)",
		chatID, userA)
	s.Require().NoError(err)

	chats, err := s.repo.GetUserChatsExtended(context.Background(), userA)
	s.Require().NoError(err)
	s.Require().Len(chats, 1)
	s.Require().Nil(chats[0].LastMessage)
	s.Require().Equal(0, chats[0].UnreadCount)
}

func (s *ChatRepositoryTestSuite) TestGetUserChatsExtended_UnreadCount() {
	userA := s.createUser("unread-user-a", "pass")
	userB := s.createUser("unread-user-b", "pass")

	var chatID int64
	err := s.pool.QueryRow(context.Background(),
		"INSERT INTO chats (name) VALUES('unread-chat') RETURNING id").Scan(&chatID)
	s.Require().NoError(err)

	_, err = s.pool.Exec(context.Background(),
		"INSERT INTO chat_members (chat_id, user_id) VALUES($1,$2),($1,$3)",
		chatID, userA, userB)
	s.Require().NoError(err)

	_, err = s.pool.Exec(context.Background(),
		"INSERT INTO messages (chat_id, sender_id, message, is_read) VALUES($1,$2,'unread-1',false),($1,$2,'unread-2',false),($1,$3,'own',false)",
		chatID, userB, userA)
	s.Require().NoError(err)

	chats, err := s.repo.GetUserChatsExtended(context.Background(), userA)
	s.Require().NoError(err)
	s.Require().Len(chats, 1)
	s.Require().Equal(2, chats[0].UnreadCount, "only userB messages should count as unread")
}
