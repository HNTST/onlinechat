package chats_postgres_repository_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	core_postgres_pool "discord/internal/core/repository/postgres/pool"
	chats_postgres_repository "discord/internal/features/chat/repository/postgres"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
)

type ChatRepositoryTestSuite struct {
	suite.Suite
	pool    *pgxpool.Pool
	wrapper core_postgres_pool.Pool
	repo    *chats_postgres_repository.ChatRepository
}

func TestChatRepositorySuite(t *testing.T) {
	suite.Run(t, new(ChatRepositoryTestSuite))
}

func (s *ChatRepositoryTestSuite) SetupSuite() {
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

	s.wrapper = &chatPoolWrapper{Pool: p}
	s.repo = chats_postgres_repository.NewChatRepository(s.wrapper)
}

func (s *ChatRepositoryTestSuite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}
}

func (s *ChatRepositoryTestSuite) SetupTest() {
	_, err := s.pool.Exec(context.Background(),
		"TRUNCATE users, chats, chat_members, messages, files RESTART IDENTITY CASCADE")
	s.Require().NoError(err)
}

func (s *ChatRepositoryTestSuite) createUser(username, password string) int64 {
	var id int64
	err := s.pool.QueryRow(context.Background(),
		"INSERT INTO users (username, password, reg) VALUES($1,$2,$3) RETURNING id",
		username, password, time.Now()).Scan(&id)
	s.Require().NoError(err)
	return id
}

func (s *ChatRepositoryTestSuite) applyMigrations(ctx context.Context) {
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
			name VARCHAR(255) NOT NULL,
			is_group BOOLEAN NOT NULL DEFAULT false
		)`,
		`CREATE TABLE IF NOT EXISTS chat_members (
			chat_id INTEGER REFERENCES chats(id) ON DELETE CASCADE,
			user_id INTEGER REFERENCES users(id),
			PRIMARY KEY (chat_id, user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS messages (
			id SERIAL PRIMARY KEY,
			chat_id INTEGER REFERENCES chats(id) ON DELETE CASCADE,
			sender_id INTEGER REFERENCES users(id),
			message TEXT NOT NULL,
			sent_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
			is_edited BOOLEAN NOT NULL DEFAULT FALSE,
			is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
			is_read BOOLEAN NOT NULL DEFAULT FALSE
		)`,
	}
	for _, m := range migrations {
		_, err := s.pool.Exec(ctx, m)
		s.Require().NoError(err)
	}
}

type chatPoolWrapper struct {
	*pgxpool.Pool
}

func (pw *chatPoolWrapper) OpTimeout() time.Duration {
	return 10 * time.Second
}

func (pw *chatPoolWrapper) GetTx(ctx context.Context) pgx.Tx {
	if tx, ok := ctx.Value("tx").(pgx.Tx); ok {
		return tx
	}
	return nil
}

func (pw *chatPoolWrapper) CtxWithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := pw.Begin(ctx)
	if err != nil {
		return fmt.Errorf("create tx: %w", err)
	}
	defer tx.Rollback(ctx)

	ctx = context.WithValue(ctx, "tx", tx)
	if err := fn(ctx); err != nil {
		return fmt.Errorf("execute fn: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
