package users_postgres_repository_test

import (
	"context"
	"time"

	"discord/internal/core/domain"
)

func (s *UserRepositoryTestSuite) TestSaveUser_Success() {
	input := domain.NewUserUninitialized("alice", "secure123")
	input.Reg = time.Now()

	user, err := s.repo.SaveUser(context.Background(), input)

	s.Require().NoError(err)
	s.Require().Greater(user.Id, int64(0))
	s.Require().Equal("alice", user.Name)
	s.Require().Equal("-1", user.Password)
	s.Require().False(user.Reg.IsZero())

	var dbUsername string
	err = s.pool.QueryRow(context.Background(),
		"SELECT username FROM users WHERE id=$1", user.Id).Scan(&dbUsername)
	s.Require().NoError(err)
	s.Require().Equal("alice", dbUsername)
}

func (s *UserRepositoryTestSuite) TestSaveUser_Duplicate() {
	input := domain.NewUserUninitialized("bob", "pass123")
	input.Reg = time.Now()

	_, err := s.repo.SaveUser(context.Background(), input)
	s.Require().NoError(err)

	_, err = s.repo.SaveUser(context.Background(), input)
	s.Require().Error(err)
}

func (s *UserRepositoryTestSuite) TestSaveUser_AvatarID_Null() {
	input := domain.NewUserUninitialized("alice", "secure123")
	input.Reg = time.Now()

	user, err := s.repo.SaveUser(context.Background(), input)

	s.Require().NoError(err)
	s.Require().Nil(user.AvatarID)
}
