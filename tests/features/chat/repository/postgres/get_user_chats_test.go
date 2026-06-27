package chats_postgres_repository_test

import (
	"context"
)

func (s *ChatRepositoryTestSuite) TestGetUserChats_Success() {
	userID := s.createUser("alice", "pass")
	otherID := s.createUser("bob", "pass")

	var chat1ID, chat2ID int64
	err := s.pool.QueryRow(context.Background(),
		"INSERT INTO chats (name) VALUES('chat-a') RETURNING id").Scan(&chat1ID)
	s.Require().NoError(err)
	err = s.pool.QueryRow(context.Background(),
		"INSERT INTO chats (name) VALUES('chat-b') RETURNING id").Scan(&chat2ID)
	s.Require().NoError(err)

	_, err = s.pool.Exec(context.Background(),
		"INSERT INTO chat_members (chat_id, user_id) VALUES($1,$2),($3,$2),($3,$4)",
		chat1ID, userID, chat2ID, otherID)
	s.Require().NoError(err)

	chats, err := s.repo.GetUserChats(context.Background(), userID)

	s.Require().NoError(err)
	s.Require().Len(chats, 2)
	s.Require().Equal("chat-a", chats[0].Name)
	s.Require().Equal("chat-b", chats[1].Name)
}

func (s *ChatRepositoryTestSuite) TestGetUserChats_NoChats() {
	userID := s.createUser("lonely", "pass")

	chats, err := s.repo.GetUserChats(context.Background(), userID)

	s.Require().NoError(err)
	s.Require().Empty(chats)
}

func (s *ChatRepositoryTestSuite) TestGetUserChats_NotFoundUser() {
	chats, err := s.repo.GetUserChats(context.Background(), 99999)

	s.Require().NoError(err)
	s.Require().Empty(chats)
}
