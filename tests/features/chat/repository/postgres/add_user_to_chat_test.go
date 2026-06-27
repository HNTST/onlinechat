package chats_postgres_repository_test

import (
	"context"
)

func (s *ChatRepositoryTestSuite) TestAddUserToChat_WithoutTx() {
	user1ID := s.createUser("eve", "pass1")
	user2ID := s.createUser("frank", "pass2")
	user3ID := s.createUser("grace", "pass3")

	var chatID int64
	err := s.pool.QueryRow(context.Background(),
		"INSERT INTO chats (name) VALUES('group') RETURNING id").Scan(&chatID)
	s.Require().NoError(err)

	err = s.repo.AddUserToChat(context.Background(), chatID, []int64{user1ID, user2ID})
	s.Require().NoError(err)

	var count int
	err = s.pool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM chat_members WHERE chat_id=$1", chatID).Scan(&count)
	s.Require().NoError(err)
	s.Require().Equal(2, count)

	err = s.repo.AddUserToChat(context.Background(), chatID, []int64{user3ID})
	s.Require().NoError(err)

	err = s.pool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM chat_members WHERE chat_id=$1", chatID).Scan(&count)
	s.Require().NoError(err)
	s.Require().Equal(3, count)
}

func (s *ChatRepositoryTestSuite) TestAddUserToChat_Duplicate() {
	user1ID := s.createUser("heidi", "pass1")
	user2ID := s.createUser("ivan", "pass2")

	var chatID int64
	err := s.pool.QueryRow(context.Background(),
		"INSERT INTO chats (name) VALUES('duo') RETURNING id").Scan(&chatID)
	s.Require().NoError(err)

	err = s.repo.AddUserToChat(context.Background(), chatID, []int64{user1ID, user2ID})
	s.Require().NoError(err)

	err = s.repo.AddUserToChat(context.Background(), chatID, []int64{user1ID})
	s.Require().NoError(err)

	var count int
	err = s.pool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM chat_members WHERE chat_id=$1", chatID).Scan(&count)
	s.Require().NoError(err)
	s.Require().Equal(2, count)
}

func (s *ChatRepositoryTestSuite) TestAddUserToChat_WithTx() {
	user1ID := s.createUser("judy", "pass1")
	user2ID := s.createUser("karl", "pass2")

	var chatID int64
	err := s.pool.QueryRow(context.Background(),
		"INSERT INTO chats (name) VALUES('tx-group') RETURNING id").Scan(&chatID)
	s.Require().NoError(err)

	ctx := context.Background()
	err = s.wrapper.CtxWithTransaction(ctx, func(txCtx context.Context) error {
		return s.repo.AddUserToChat(txCtx, chatID, []int64{user1ID, user2ID})
	})
	s.Require().NoError(err)

	var count int
	err = s.pool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM chat_members WHERE chat_id=$1", chatID).Scan(&count)
	s.Require().NoError(err)
	s.Require().Equal(2, count)
}
