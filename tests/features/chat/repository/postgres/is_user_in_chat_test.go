package chats_postgres_repository_test

import (
	"context"
)

func (s *ChatRepositoryTestSuite) TestIsUserInChat_True() {
	userID := s.createUser("member-user", "pass")
	chatID, err := s.repo.CreateNewChat(context.Background(), "members-test", true)
	s.Require().NoError(err)

	err = s.repo.AddUserToChat(context.Background(), chatID, []int64{userID})
	s.Require().NoError(err)

	isMember, err := s.repo.IsUserInChat(context.Background(), userID, chatID)
	s.Require().NoError(err)
	s.Require().True(isMember)
}

func (s *ChatRepositoryTestSuite) TestIsUserInChat_False() {
	userID := s.createUser("non-member", "pass")
	chatID, err := s.repo.CreateNewChat(context.Background(), "no-access", true)
	s.Require().NoError(err)

	isMember, err := s.repo.IsUserInChat(context.Background(), userID, chatID)
	s.Require().NoError(err)
	s.Require().False(isMember)
}

func (s *ChatRepositoryTestSuite) TestIsUserInChat_NonExistentChat() {
	userID := s.createUser("ghost-user", "pass")

	isMember, err := s.repo.IsUserInChat(context.Background(), userID, 99999)
	s.Require().NoError(err)
	s.Require().False(isMember)
}
