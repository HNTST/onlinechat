package users_postgres_repository_test

import (
	"context"
	"time"

	"discord/internal/core/domain"
)

func (s *UserRepositoryTestSuite) TestGetUser_Success() {
	input := domain.NewUserUninitialized("charlie", "mypass")
	input.Reg = time.Now()
	saved, err := s.repo.SaveUser(context.Background(), input)
	s.Require().NoError(err)

	user, err := s.repo.GetUser(context.Background(), saved.Id)

	s.Require().NoError(err)
	s.Require().Equal("charlie", user.Name)
	s.Require().Equal("-1", user.Password)
	s.Require().False(user.Reg.IsZero())
}

func (s *UserRepositoryTestSuite) TestGetUser_NotFound() {
	_, err := s.repo.GetUser(context.Background(), 99999)

	s.Require().Error(err)
}

func (s *UserRepositoryTestSuite) TestGetUser_AvatarID() {
	// create a user
	input := domain.NewUserUninitialized("avatar-user", "pass")
	input.Reg = time.Now()
	saved, err := s.repo.SaveUser(context.Background(), input)
	s.Require().NoError(err)
	s.Require().Nil(saved.AvatarID)

	// insert a file and set avatar_id via SQL
	var fileID int64
	err = s.pool.QueryRow(context.Background(),
		`INSERT INTO files (file_key, original_file_name, type_file, size_file)
		 VALUES('avatars/test.png', 'test.png', 'image/png', 1024) RETURNING id`,
	).Scan(&fileID)
	s.Require().NoError(err)

	_, err = s.pool.Exec(context.Background(),
		"UPDATE users SET avatar_id = $1 WHERE id = $2", fileID, saved.Id)
	s.Require().NoError(err)

	user, err := s.repo.GetUser(context.Background(), saved.Id)

	s.Require().NoError(err)
	s.Require().NotNil(user.AvatarID)
	s.Require().Equal(fileID, *user.AvatarID)
}
