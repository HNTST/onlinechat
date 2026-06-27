package chats_postgres_repository_test

import (
	"context"
)

func (s *ChatRepositoryTestSuite) TestCheckExistsChat_Found() {
	user1ID := s.createUser("alice", "pass1")
	user2ID := s.createUser("bob", "pass2")

	var chatID int64
	err := s.pool.QueryRow(context.Background(),
		"INSERT INTO chats (name) VALUES('duo') RETURNING id").Scan(&chatID)
	s.Require().NoError(err)

	_, err = s.pool.Exec(context.Background(),
		"INSERT INTO chat_members (chat_id, user_id) VALUES($1,$2),($1,$3)",
		chatID, user1ID, user2ID)
	s.Require().NoError(err)

	foundID, err := s.repo.CheckExistsChat(context.Background(), user1ID, user2ID)

	s.Require().NoError(err)
	s.Require().Equal(chatID, foundID)
}

func (s *ChatRepositoryTestSuite) TestCheckExistsChat_NotFound() {
	user1ID := s.createUser("charlie", "pass1")
	user2ID := s.createUser("dave", "pass2")

	_, err := s.repo.CheckExistsChat(context.Background(), user1ID, user2ID)

	s.Require().Error(err)
}
