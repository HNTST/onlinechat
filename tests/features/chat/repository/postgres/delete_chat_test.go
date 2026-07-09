package chats_postgres_repository_test

import (
	"context"
)

func (s *ChatRepositoryTestSuite) TestDeleteChat_Success() {
	chatID, err := s.repo.CreateNewChat(context.Background(), "to-delete", true)
	s.Require().NoError(err)

	var count int
	err = s.pool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM chats WHERE id=$1", chatID).Scan(&count)
	s.Require().NoError(err)
	s.Require().Equal(1, count)

	err = s.repo.DeleteChat(context.Background(), chatID)
	s.Require().NoError(err)

	err = s.pool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM chats WHERE id=$1", chatID).Scan(&count)
	s.Require().NoError(err)
	s.Require().Equal(0, count)
}

func (s *ChatRepositoryTestSuite) TestDeleteChat_CascadesMembers() {
	userID := s.createUser("delete-test-user", "pass")
	chatID, err := s.repo.CreateNewChat(context.Background(), "cascade-test", true)
	s.Require().NoError(err)

	err = s.repo.AddUserToChat(context.Background(), chatID, []int64{userID})
	s.Require().NoError(err)

	err = s.repo.DeleteChat(context.Background(), chatID)
	s.Require().NoError(err)

	var memberCount int
	err = s.pool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM chat_members WHERE chat_id=$1", chatID).Scan(&memberCount)
	s.Require().NoError(err)
	s.Require().Equal(0, memberCount, "chat_members must cascade on chat delete")
}
