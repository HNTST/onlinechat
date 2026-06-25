package postgres_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"discord/internal/core/domain"
	core_postgres_pool "discord/internal/core/repository/postgres/pool"
	oldrepo "discord/internal/repository/postgres"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"
)

type OldRepositoryTestSuite struct {
	suite.Suite
	pool    *pgxpool.Pool
	wrapper core_postgres_pool.Pool
	repo    *oldrepo.OldDbREpository
}

func TestOldRepositorySuite(t *testing.T) {
	suite.Run(t, new(OldRepositoryTestSuite))
}

func (s *OldRepositoryTestSuite) SetupSuite() {
	host := getEnv("TEST_DB_HOST", "127.0.0.1")
	port := getEnv("TEST_DB_PORT", "5432")
	user := getEnv("TEST_DB_USER", "discord_user")
	password := getEnv("TEST_DB_PASSWORD", "password")
	dbname := getEnv("TEST_DB_NAME", "discord_test")

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, password, host, port, dbname)

	ctx := context.Background()
	p, err := pgxpool.New(ctx, connStr)
	s.Require().NoError(err, "failed to connect to test database")

	err = p.Ping(ctx)
	s.Require().NoError(err, "failed to ping test database")

	s.pool = p

	s.applyMigrations(ctx)

	s.wrapper = &poolWrapper{Pool: p}
	s.repo = oldrepo.NewOldRepostiroty(s.wrapper)
}

func (s *OldRepositoryTestSuite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}
}

func (s *OldRepositoryTestSuite) SetupTest() {
	ctx := context.Background()
	_, err := s.pool.Exec(ctx, "TRUNCATE users, chats, chat_members, messages, files RESTART IDENTITY CASCADE")
	s.Require().NoError(err, "failed to truncate tables")
}

// --- helpers

func (s *OldRepositoryTestSuite) createUser(username, password string) int64 {
	ctx := context.Background()
	id, err := s.repo.SaveUser(ctx, username, password)
	s.Require().NoError(err)
	return id
}

func (s *OldRepositoryTestSuite) directCreateChat(name string, memberIDs ...int64) int64 {
	ctx := context.Background()
	var chatID int64
	err := s.pool.QueryRow(ctx, "INSERT INTO chats (name) VALUES($1) RETURNING id", name).Scan(&chatID)
	s.Require().NoError(err)
	for _, uid := range memberIDs {
		_, err = s.pool.Exec(ctx, "INSERT INTO chat_members (chat_id, user_id) VALUES($1, $2) ON CONFLICT DO NOTHING", chatID, uid)
		s.Require().NoError(err)
	}
	return chatID
}

// --- SaveUser

func (s *OldRepositoryTestSuite) TestSaveUser_Success() {
	id, err := s.repo.SaveUser(context.Background(), "alice", "secret123")

	s.Require().NoError(err)
	s.Require().Greater(id, int64(0))

	var username string
	err = s.pool.QueryRow(context.Background(), "SELECT username FROM users WHERE id=$1", id).Scan(&username)
	s.Require().NoError(err)
	s.Require().Equal("alice", username)
}

func (s *OldRepositoryTestSuite) TestSaveUser_DuplicateUsername() {
	ctx := context.Background()

	_, err := s.repo.SaveUser(ctx, "alice", "secret123")
	s.Require().NoError(err)

	_, err = s.repo.SaveUser(ctx, "alice", "otherpass")
	s.Require().Error(err)
}

// --- GetUser

func (s *OldRepositoryTestSuite) TestGetUser_Success() {
	id := s.createUser("bob", "pass123")

	user, err := s.repo.GetUser(context.Background(), id)

	s.Require().NoError(err)
	s.Require().Equal("bob", user.Name)
	s.Require().False(user.Reg.IsZero())
}

func (s *OldRepositoryTestSuite) TestGetUser_NotFound() {
	_, err := s.repo.GetUser(context.Background(), 99999)
	s.Require().Error(err)
}

// --- GetUserID

func (s *OldRepositoryTestSuite) TestGetUserID_Success() {
	id := s.createUser("charlie", "mypass")

	foundID, err := s.repo.GetUserID(context.Background(), "charlie")

	s.Require().NoError(err)
	s.Require().Equal(id, foundID)
}

func (s *OldRepositoryTestSuite) TestGetUserID_NotFound() {
	_, err := s.repo.GetUserID(context.Background(), "nonexistent")
	s.Require().Error(err)
}

// --- SaveChatWithAuthor

func (s *OldRepositoryTestSuite) TestSaveChatWithAuthor_CreatesNewChat() {
	ctx := context.Background()

	user1ID := s.createUser("dave", "pass1")
	user2ID := s.createUser("eve", "pass2")

	chatID, err := s.repo.SaveChatWithAuthor(ctx, "dave-eve-chat", user1ID, user2ID)

	s.Require().NoError(err)
	s.Require().Greater(chatID, int64(0))

	var memberCount int
	err = s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM chat_members WHERE chat_id=$1", chatID).Scan(&memberCount)
	s.Require().NoError(err)
	s.Require().Equal(2, memberCount)

	var chatName string
	err = s.pool.QueryRow(ctx, "SELECT name FROM chats WHERE id=$1", chatID).Scan(&chatName)
	s.Require().NoError(err)
	s.Require().Equal("dave-eve-chat", chatName)
}

func (s *OldRepositoryTestSuite) TestSaveChatWithAuthor_ReturnsExistingChat() {
	ctx := context.Background()

	user1ID := s.createUser("frank", "pass1")
	user2ID := s.createUser("grace", "pass2")

	_ = s.directCreateChat("existing-chat", user1ID, user2ID)

	chatID, err := s.repo.SaveChatWithAuthor(ctx, "attempt-new-name", user1ID, user2ID)

	s.Require().NoError(err)
	s.Require().Greater(chatID, int64(0))

	var chatName string
	err = s.pool.QueryRow(ctx, "SELECT name FROM chats WHERE id=$1", chatID).Scan(&chatName)
	s.Require().NoError(err)
	s.Require().Equal("existing-chat", chatName)
}

// --- GetUserChats

func (s *OldRepositoryTestSuite) TestGetUserChats() {
	user1ID := s.createUser("heidi", "pass1")
	user2ID := s.createUser("ivan", "pass2")
	user3ID := s.createUser("judy", "pass3")

	_ = s.directCreateChat("chat-ab", user1ID, user2ID)
	_ = s.directCreateChat("chat-ac", user1ID, user3ID)

	chats, err := s.repo.GetUserChats(context.Background(), user1ID)

	s.Require().NoError(err)
	s.Require().Len(chats, 2)

	names := make(map[string]bool)
	for _, c := range chats {
		names[c.Name] = true
	}
	s.Require().True(names["chat-ab"])
	s.Require().True(names["chat-ac"])

	chatsUser2, err := s.repo.GetUserChats(context.Background(), user2ID)
	s.Require().NoError(err)
	s.Require().Len(chatsUser2, 1)
	s.Require().Equal("chat-ab", chatsUser2[0].Name)

	chatsUserNone, err := s.repo.GetUserChats(context.Background(), 99999)
	s.Require().NoError(err)
	s.Require().Len(chatsUserNone, 0)
}

// --- GetChatMembers

func (s *OldRepositoryTestSuite) TestGetChatMembers() {
	user1ID := s.createUser("karen", "pass1")
	user2ID := s.createUser("leo", "pass2")

	chatID := s.directCreateChat("test-chat", user1ID, user2ID)

	members, err := s.repo.GetChatMembers(context.Background(), chatID)

	s.Require().NoError(err)
	s.Require().Len(members, 2)
	s.Require().Contains(members, user1ID)
	s.Require().Contains(members, user2ID)
}

func (s *OldRepositoryTestSuite) TestGetChatMembers_None() {
	user1ID := s.createUser("nobody", "pass1")
	user2ID := s.createUser("nowhere", "pass2")
	chatID := s.directCreateChat("empty-members-chat", user1ID, user2ID)

	err := s.pool.QueryRow(context.Background(), "DELETE FROM chat_members WHERE chat_id=$1", chatID).Scan()
	s.Require().Error(err)

	members, err := s.repo.GetChatMembers(context.Background(), chatID)
	s.Require().NoError(err)
	s.Require().Len(members, 0)
}

// --- SaveUserToChat

func (s *OldRepositoryTestSuite) TestSaveUserToChat() {
	user1ID := s.createUser("mallory", "pass1")
	user2ID := s.createUser("nancy", "pass2")
	user3ID := s.createUser("oscar", "pass3")

	chatID := s.directCreateChat("group-chat", user1ID, user2ID)

	err := s.repo.SaveUserToChat(context.Background(), user3ID, chatID)
	s.Require().NoError(err)

	members, err := s.repo.GetChatMembers(context.Background(), chatID)
	s.Require().NoError(err)
	s.Require().Len(members, 3)
	s.Require().Contains(members, user3ID)
}

func (s *OldRepositoryTestSuite) TestSaveUserToChat_Duplicate() {
	user1ID := s.createUser("u1", "pass1")
	user2ID := s.createUser("u2", "pass2")

	chatID := s.directCreateChat("duo-chat", user1ID, user2ID)

	err := s.repo.SaveUserToChat(context.Background(), user1ID, chatID)
	s.Require().Error(err)
}

// --- SaveMessage + GetChatMessages

func (s *OldRepositoryTestSuite) TestSaveMessageAndGetChatMessages() {
	user1ID := s.createUser("peggy", "pass1")
	user2ID := s.createUser("quentin", "pass2")

	chatID := s.directCreateChat("msg-chat", user1ID, user2ID)

	now := time.Now().UTC()
	msg1 := domain.Messages{
		ChatId:   chatID,
		SenderId: user1ID,
		Message:  "Hello!",
		Sent_at:  now,
	}
	err := s.repo.SaveMessage(context.Background(), msg1)
	s.Require().NoError(err)

	msg2 := domain.Messages{
		ChatId:   chatID,
		SenderId: user2ID,
		Message:  "Hi there!",
		Sent_at:  now.Add(time.Second),
	}
	err = s.repo.SaveMessage(context.Background(), msg2)
	s.Require().NoError(err)

	messages, err := s.repo.GetChatMessages(context.Background(), chatID)

	s.Require().NoError(err)
	s.Require().Len(messages, 2)
	s.Require().Equal(user1ID, messages[0].SenderId)
	s.Require().Equal("Hello!", messages[0].Message)
	s.Require().Equal(user2ID, messages[1].SenderId)
	s.Require().Equal("Hi there!", messages[1].Message)
}

func (s *OldRepositoryTestSuite) TestGetChatMessages_Empty() {
	user1ID := s.createUser("r1", "pass1")
	user2ID := s.createUser("r2", "pass2")
	chatID := s.directCreateChat("empty-chat", user1ID, user2ID)

	messages, err := s.repo.GetChatMessages(context.Background(), chatID)

	s.Require().NoError(err)
	s.Require().Len(messages, 0)
}

// --- FindExistingChat

func (s *OldRepositoryTestSuite) TestFindExistingChat_Found() {
	user1ID := s.createUser("romeo", "pass1")
	user2ID := s.createUser("juliet", "pass2")

	chatID := s.directCreateChat("love-chat", user1ID, user2ID)

	foundID, err := s.repo.FindExistingChat(context.Background(), user1ID, user2ID)

	s.Require().NoError(err)
	s.Require().Equal(chatID, foundID)
}

func (s *OldRepositoryTestSuite) TestFindExistingChat_NotFound() {
	user1ID := s.createUser("xavier", "pass1")
	user2ID := s.createUser("yasmine", "pass2")

	_, err := s.repo.FindExistingChat(context.Background(), user1ID, user2ID)

	s.Require().Error(err)
}

// --- UpdateUserAvatarPath

func (s *OldRepositoryTestSuite) TestUpdateUserAvatarPath() {
	userID := s.createUser("avatar-user", "pass123")

	fileID, err := s.repo.UpdateUserAvatarPath(context.Background(), userID, 1024, "avatars/user_1/face.png", "face.png", "image/png")

	s.Require().NoError(err)
	s.Require().Greater(fileID, int64(0))

	var avatarID *int64
	err = s.pool.QueryRow(context.Background(), "SELECT avatar_id FROM users WHERE id=$1", userID).Scan(&avatarID)
	s.Require().NoError(err)
	s.Require().NotNil(avatarID)
	s.Require().Equal(fileID, *avatarID)

	var key, origName, fileType string
	var fileSize int64
	err = s.pool.QueryRow(context.Background(), "SELECT file_key, original_file_name, type_file, size_file FROM files WHERE id=$1", fileID).
		Scan(&key, &origName, &fileType, &fileSize)
	s.Require().NoError(err)
	s.Require().Equal("avatars/user_1/face.png", key)
	s.Require().Equal("face.png", origName)
	s.Require().Equal("image/png", fileType)
	s.Require().Equal(int64(1024), fileSize)
}

// --- migrations

func (s *OldRepositoryTestSuite) applyMigrations(ctx context.Context) {
	migrations := []string{
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
		`CREATE TABLE IF NOT EXISTS files (
			id SERIAL PRIMARY KEY,
			file_key VARCHAR(255) UNIQUE NOT NULL,
			original_file_name VARCHAR(255) NOT NULL,
			type_file VARCHAR(255) NOT NULL,
			size_file BIGINT NOT NULL
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
		s.Require().NoError(err, "migration failed")
	}
}

// poolWrapper adapts *pgxpool.Pool to the project's Pool interface

type poolWrapper struct {
	*pgxpool.Pool
}

func (pw *poolWrapper) OpTimeout() time.Duration {
	return 10 * time.Second
}

func (pw *poolWrapper) GetTx(_ context.Context) pgx.Tx {
	return nil
}

func (pw *poolWrapper) CtxWithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
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
