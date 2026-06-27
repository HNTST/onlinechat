package users_postgres_repository_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	core_postgres_pool "discord/internal/core/repository/postgres/pool"
	users_postgres_repository "discord/internal/features/user/repository/postgres"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
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
	_ = godotenv.Load("../../../../internal/core/config/.env.test")

	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "test_user")
	password := getEnv("DB_PASSWORD", "test_pass")
	dbname := getEnv("DB_NAME", "discord_test")

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
			reg TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			avatar_id INTEGER REFERENCES files(id) ON DELETE SET NULL
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
