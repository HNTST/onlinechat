package chats_postgres_repository_test

import (
	"context"
)

func (s *ChatRepositoryTestSuite) TestUpdateChatName_Success() {
	chatID, err := s.repo.CreateNewChat(context.Background(), "old-name", false)
	s.Require().NoError(err)

	err = s.repo.UpdateChatName(context.Background(), chatID, "new-name")
	s.Require().NoError(err)

	var name string
	err = s.pool.QueryRow(context.Background(),
		"SELECT name FROM chats WHERE id=$1", chatID).Scan(&name)
	s.Require().NoError(err)
	s.Require().Equal("new-name", name)
}

func (s *ChatRepositoryTestSuite) TestUpdateChatName_WithTx() {
	ctx := context.Background()

	chatID, err := s.repo.CreateNewChat(ctx, "tx-old-name", false)
	s.Require().NoError(err)

	err = s.wrapper.CtxWithTransaction(ctx, func(txCtx context.Context) error {
		return s.repo.UpdateChatName(txCtx, chatID, "tx-new-name")
	})
	s.Require().NoError(err)

	var name string
	err = s.pool.QueryRow(ctx,
		"SELECT name FROM chats WHERE id=$1", chatID).Scan(&name)
	s.Require().NoError(err)
	s.Require().Equal("tx-new-name", name)
}
