package chats_postgres_repository_test

import (
	"context"
)

func (s *ChatRepositoryTestSuite) TestCreateNewChat_WithoutTx() {
	chatID, err := s.repo.CreateNewChat(context.Background(), "test-chat", false)

	s.Require().NoError(err)
	s.Require().Greater(chatID, int64(0))

	var name string
	err = s.pool.QueryRow(context.Background(),
		"SELECT name FROM chats WHERE id=$1", chatID).Scan(&name)
	s.Require().NoError(err)
	s.Require().Equal("test-chat", name)
}

func (s *ChatRepositoryTestSuite) TestCreateNewChat_WithIsGroup() {
	chatID, err := s.repo.CreateNewChat(context.Background(), "group-chat", true)

	s.Require().NoError(err)
	s.Require().Greater(chatID, int64(0))

	var name string
	var isGroup bool
	err = s.pool.QueryRow(context.Background(),
		"SELECT name, is_group FROM chats WHERE id=$1", chatID).Scan(&name, &isGroup)
	s.Require().NoError(err)
	s.Require().Equal("group-chat", name)
	s.Require().True(isGroup)
}

func (s *ChatRepositoryTestSuite) TestCreateNewChat_WithTx() {
	ctx := context.Background()

	var chatID int64
	err := s.wrapper.CtxWithTransaction(ctx, func(txCtx context.Context) error {
		var err error
		chatID, err = s.repo.CreateNewChat(txCtx, "tx-chat", false)
		s.Require().NoError(err)
		s.Require().Greater(chatID, int64(0))
		return nil
	})
	s.Require().NoError(err)

	var name string
	err = s.pool.QueryRow(context.Background(),
		"SELECT name FROM chats WHERE id=$1", chatID).Scan(&name)
	s.Require().NoError(err)
	s.Require().Equal("tx-chat", name)
}
