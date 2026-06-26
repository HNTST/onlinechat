package chats_postgres_repository_test

import (
	"context"
	"discord/internal/core/domain"
	"fmt"
	"os"
	"testing"
	"time"

	core_postgres_pool "discord/internal/core/repository/postgres/pool"
	chats_postgres_repository "discord/internal/features/chat/repository/postgres"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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

// --- CreateNewChat without transaction

func (s *ChatRepositoryTestSuite) TestCreateNewChat_WithoutTx() {
	chatID, err := s.repo.CreateNewChat(context.Background(), "test-chat")

	s.Require().NoError(err)
	s.Require().Greater(chatID, int64(0))

	var name string
	err = s.pool.QueryRow(context.Background(),
		"SELECT name FROM chats WHERE id=$1", chatID).Scan(&name)
	s.Require().NoError(err)
	s.Require().Equal("test-chat", name)
}

// --- CreateNewChat with transaction

func (s *ChatRepositoryTestSuite) TestCreateNewChat_WithTx() {
	ctx := context.Background()

	var chatID int64
	err := s.wrapper.CtxWithTransaction(ctx, func(txCtx context.Context) error {
		var err error
		chatID, err = s.repo.CreateNewChat(txCtx, "tx-chat")
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

// --- CheckExistsChat found

func (s *ChatRepositoryTestSuite) TestCheckExistsChat_Found() {
	user1ID := s.createUser("alice", "pass1")
	user2ID := s.createUser("bob", "pass2")

	var chatID int64
	err := s.pool.QueryRow(context.Background(),
		"INSERT INTO chats (name) VALUES('duo') RETURNING id").Scan(&chatID)
	s.Require().NoError(err)

	_, err = s.pool.Exec(context.Background(),
		"INSERT INTO chat_members (chat_id, user_id) VALUES($1,$2),($1,$3)",
		chatID, user1ID, user2ID)
	s.Require().NoError(err)

	foundID, err := s.repo.CheckExistsChat(context.Background(), user1ID, user2ID)

	s.Require().NoError(err)
	s.Require().Equal(chatID, foundID)
}

// --- CheckExistsChat not found

func (s *ChatRepositoryTestSuite) TestCheckExistsChat_NotFound() {
	user1ID := s.createUser("charlie", "pass1")
	user2ID := s.createUser("dave", "pass2")

	_, err := s.repo.CheckExistsChat(context.Background(), user1ID, user2ID)

	s.Require().Error(err)
}

// --- AddUserToChat without transaction

func (s *ChatRepositoryTestSuite) TestAddUserToChat_WithoutTx() {
	user1ID := s.createUser("eve", "pass1")
	user2ID := s.createUser("frank", "pass2")
	user3ID := s.createUser("grace", "pass3")

	var chatID int64
	err := s.pool.QueryRow(context.Background(),
		"INSERT INTO chats (name) VALUES('group') RETURNING id").Scan(&chatID)
	s.Require().NoError(err)

	err = s.repo.AddUserToChat(context.Background(), chatID, []int64{user1ID, user2ID})
	s.Require().NoError(err)

	var count int
	err = s.pool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM chat_members WHERE chat_id=$1", chatID).Scan(&count)
	s.Require().NoError(err)
	s.Require().Equal(2, count)

	err = s.repo.AddUserToChat(context.Background(), chatID, []int64{user3ID})
	s.Require().NoError(err)

	err = s.pool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM chat_members WHERE chat_id=$1", chatID).Scan(&count)
	s.Require().NoError(err)
	s.Require().Equal(3, count)
}

// --- AddUserToChat duplicate (on conflict do nothing)

func (s *ChatRepositoryTestSuite) TestAddUserToChat_Duplicate() {
	user1ID := s.createUser("heidi", "pass1")
	user2ID := s.createUser("ivan", "pass2")

	var chatID int64
	err := s.pool.QueryRow(context.Background(),
		"INSERT INTO chats (name) VALUES('duo') RETURNING id").Scan(&chatID)
	s.Require().NoError(err)

	err = s.repo.AddUserToChat(context.Background(), chatID, []int64{user1ID, user2ID})
	s.Require().NoError(err)

	err = s.repo.AddUserToChat(context.Background(), chatID, []int64{user1ID})
	s.Require().NoError(err)

	var count int
	err = s.pool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM chat_members WHERE chat_id=$1", chatID).Scan(&count)
	s.Require().NoError(err)
	s.Require().Equal(2, count)
}

// --- AddUserToChat with transaction

func (s *ChatRepositoryTestSuite) TestAddUserToChat_WithTx() {
	user1ID := s.createUser("judy", "pass1")
	user2ID := s.createUser("karl", "pass2")

	var chatID int64
	err := s.pool.QueryRow(context.Background(),
		"INSERT INTO chats (name) VALUES('tx-group') RETURNING id").Scan(&chatID)
	s.Require().NoError(err)

	ctx := context.Background()
	err = s.wrapper.CtxWithTransaction(ctx, func(txCtx context.Context) error {
		return s.repo.AddUserToChat(txCtx, chatID, []int64{user1ID, user2ID})
	})
	s.Require().NoError(err)

	var count int
	err = s.pool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM chat_members WHERE chat_id=$1", chatID).Scan(&count)
	s.Require().NoError(err)
	s.Require().Equal(2, count)
}

// --- GetUserChats

func (s *ChatRepositoryTestSuite) TestGetUserChats_Success() {
	userID := s.createUser("alice", "pass")
	otherID := s.createUser("bob", "pass")

	var chat1ID, chat2ID int64
	err := s.pool.QueryRow(context.Background(),
		"INSERT INTO chats (name) VALUES('chat-a') RETURNING id").Scan(&chat1ID)
	s.Require().NoError(err)
	err = s.pool.QueryRow(context.Background(),
		"INSERT INTO chats (name) VALUES('chat-b') RETURNING id").Scan(&chat2ID)
	s.Require().NoError(err)

	_, err = s.pool.Exec(context.Background(),
		"INSERT INTO chat_members (chat_id, user_id) VALUES($1,$2),($3,$2),($3,$4)",
		chat1ID, userID, chat2ID, otherID)
	s.Require().NoError(err)

	chats, err := s.repo.GetUserChats(context.Background(), userID)

	s.Require().NoError(err)
	s.Require().Len(chats, 2)
	s.Require().Equal("chat-a", chats[0].Name)
	s.Require().Equal("chat-b", chats[1].Name)
}

func (s *ChatRepositoryTestSuite) TestGetUserChats_NoChats() {
	userID := s.createUser("lonely", "pass")

	chats, err := s.repo.GetUserChats(context.Background(), userID)

	s.Require().NoError(err)
	s.Require().Empty(chats)
}

// --- GetChatMessages

func (s *ChatRepositoryTestSuite) TestGetChatMessages_Success() {
	userID := s.createUser("msg-sender", "pass")

	var chatID int64
	err := s.pool.QueryRow(context.Background(),
		"INSERT INTO chats (name) VALUES('msg-chat') RETURNING id").Scan(&chatID)
	s.Require().NoError(err)

	_, err = s.pool.Exec(context.Background(),
		"INSERT INTO messages (chat_id, sender_id, message) VALUES($1,$2,'hi'),($1,$2,'there')",
		chatID, userID)
	s.Require().NoError(err)

	msgs, err := s.repo.GetChatMessages(context.Background(), chatID)

	s.Require().NoError(err)
	s.Require().Len(msgs, 2)
	s.Require().Equal("hi", msgs[0].Message)
	s.Require().Equal("there", msgs[1].Message)
}

func (s *ChatRepositoryTestSuite) TestGetChatMessages_NoMessages() {
	var chatID int64
	err := s.pool.QueryRow(context.Background(),
		"INSERT INTO chats (name) VALUES('empty-chat') RETURNING id").Scan(&chatID)
	s.Require().NoError(err)

	msgs, err := s.repo.GetChatMessages(context.Background(), chatID)

	s.Require().NoError(err)
	s.Require().Empty(msgs)
}

// --- GetChatMembers

func (s *ChatRepositoryTestSuite) TestGetChatMembers_Success() {
	user1ID := s.createUser("member1", "pass")
	user2ID := s.createUser("member2", "pass")

	var chatID int64
	err := s.pool.QueryRow(context.Background(),
		"INSERT INTO chats (name) VALUES('member-chat') RETURNING id").Scan(&chatID)
	s.Require().NoError(err)

	_, err = s.pool.Exec(context.Background(),
		"INSERT INTO chat_members (chat_id, user_id) VALUES($1,$2),($1,$3)",
		chatID, user1ID, user2ID)
	s.Require().NoError(err)

	members, err := s.repo.GetChatMembers(context.Background(), chatID)

	s.Require().NoError(err)
	s.Require().Len(members, 2)
	s.Require().Equal(user1ID, members[0])
	s.Require().Equal(user2ID, members[1])
}

func (s *ChatRepositoryTestSuite) TestGetChatMembers_NoMembers() {
	var chatID int64
	err := s.pool.QueryRow(context.Background(),
		"INSERT INTO chats (name) VALUES('lonely-chat') RETURNING id").Scan(&chatID)
	s.Require().NoError(err)

	members, err := s.repo.GetChatMembers(context.Background(), chatID)

	s.Require().NoError(err)
	s.Require().Empty(members)
}

// --- SaveMessage

func (s *ChatRepositoryTestSuite) TestSaveMessage_Success() {
	userID := s.createUser("sender", "pass")

	var chatID int64
	err := s.pool.QueryRow(context.Background(),
		"INSERT INTO chats (name) VALUES('save-msg-chat') RETURNING id").Scan(&chatID)
	s.Require().NoError(err)

	msg := domain.Messages{
		ChatId:   chatID,
		SenderId: userID,
		Message:  "hello world",
	}
	err = s.repo.SaveMessage(context.Background(), msg)

	s.Require().NoError(err)

	var count int
	err = s.pool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM messages WHERE chat_id=$1", chatID).Scan(&count)
	s.Require().NoError(err)
	s.Require().Equal(1, count)
}

// --- helpers

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
