package users_postgres_repository_test

import (
	"context"
	"time"

	"discord/internal/core/domain"
)

func (s *UserRepositoryTestSuite) TestGetUserID_Success() {
	input := domain.NewUserUninitialized("dave", "pass123")
	input.Reg = time.Now()
	saved, err := s.repo.SaveUser(context.Background(), input)
	s.Require().NoError(err)

	id, err := s.repo.GetUserID(context.Background(), "dave")

	s.Require().NoError(err)
	s.Require().Equal(saved.Id, id)
}

func (s *UserRepositoryTestSuite) TestGetUserID_NotFound() {
	_, err := s.repo.GetUserID(context.Background(), "nonexistent")

	s.Require().Error(err)
}
