package users_postgres_repository_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"discord/internal/core/domain"
	core_postgres_pool "discord/internal/core/repository/postgres/pool"
	users_postgres_repository "discord/internal/features/user/repository/postgres"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"
)

type UserRepositoryTestSuite struct {
	suite.Suite
	pool    *pgxpool.Pool
	wrapper core_postgres_pool.Pool
	repo    *users_postgres_repository.UserRepository
}

func TestUserRepositorySuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}

func (s *UserRepositoryTestSuite) SetupSuite() {
	host := getEnv("TEST_DB_HOST", "127.0.0.1")
	port := getEnv("TEST_DB_PORT", "5432")
	user := getEnv("TEST_DB_USER", "discord_user")
	password := getEnv("TEST_DB_PASSWORD", "password")
	dbname := getEnv("TEST_DB_NAME", "discord_test")

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, password, host, port, dbname)

	ctx := context.Background()
	p, err := pgxpool.New(ctx, connStr)
	s.Require().NoError(err)
	err = p.Ping(ctx)
	s.Require().NoError(err)

	s.pool = p
	s.applyMigrations(ctx)

	s.wrapper = &userPoolWrapper{Pool: p}
	s.repo = users_postgres_repository.NewUsersRepository(s.wrapper)
}

func (s *UserRepositoryTestSuite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}
}

func (s *UserRepositoryTestSuite) SetupTest() {
	_, err := s.pool.Exec(context.Background(),
		"TRUNCATE users, chats, chat_members, messages, files RESTART IDENTITY CASCADE")
	s.Require().NoError(err)
}

// --- SaveUser

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

// --- GetUser

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

// --- GetUserID

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

// --- helpers

func (s *UserRepositoryTestSuite) applyMigrations(ctx context.Context) {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS files (
			id SERIAL PRIMARY KEY,
			file_key VARCHAR(255) UNIQUE NOT NULL,
			original_file_name VARCHAR(255) NOT NULL,
			type_file VARCHAR(255) NOT NULL,
			size_file BIGINT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(255) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			reg TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS chats (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS chat_members (
			chat_id INTEGER REFERENCES chats(id),
			user_id INTEGER REFERENCES users(id),
			PRIMARY KEY (chat_id, user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS messages (
			id SERIAL PRIMARY KEY,
			chat_id INTEGER REFERENCES chats(id) ON DELETE CASCADE,
			sender_id INTEGER REFERENCES users(id),
			message TEXT NOT NULL,
			sent_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`DO $$ BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name='users' AND column_name='avatar_id'
			) THEN
				ALTER TABLE users ADD COLUMN avatar_id INTEGER;
			END IF;
		END $$`,
		`DO $$ BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.table_constraints
				WHERE constraint_name='fk_avatar_file' AND table_name='users'
			) THEN
				ALTER TABLE users ADD CONSTRAINT fk_avatar_file FOREIGN KEY (avatar_id) REFERENCES files(id);
			END IF;
		END $$`,
	}
	for _, m := range migrations {
		_, err := s.pool.Exec(ctx, m)
		s.Require().NoError(err)
	}
}

type userPoolWrapper struct {
	*pgxpool.Pool
}

func (pw *userPoolWrapper) OpTimeout() time.Duration {
	return 10 * time.Second
}

func (pw *userPoolWrapper) GetTx(_ context.Context) pgx.Tx {
	return nil
}

func (pw *userPoolWrapper) CtxWithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
