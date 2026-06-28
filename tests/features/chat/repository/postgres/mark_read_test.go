package chats_postgres_repository_test

import (
	"context"
)

func (s *ChatRepositoryTestSuite) TestMarkMessagesAsRead_Success() {
	userA := s.createUser("user-a", "pass")
	userB := s.createUser("user-b", "pass")

	var chatID int64
	err := s.pool.QueryRow(context.Background(),
		"INSERT INTO chats (name) VALUES('read-chat') RETURNING id").Scan(&chatID)
	s.Require().NoError(err)

	_, err = s.pool.Exec(context.Background(),
		"INSERT INTO messages (chat_id, sender_id, message, is_read) VALUES($1,$2,'msg1',false),($1,$2,'msg2',false),($1,$3,'msg3',false)",
		chatID, userA, userB)
	s.Require().NoError(err)

	var lastMsgID int64
	err = s.pool.QueryRow(context.Background(),
		"SELECT id FROM messages WHERE chat_id=$1 ORDER BY id DESC LIMIT 1", chatID).Scan(&lastMsgID)
	s.Require().NoError(err)

	err = s.repo.MarkMessagesAsRead(context.Background(), chatID, userA, lastMsgID)
	s.Require().NoError(err)

	var readCount int
	err = s.pool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM messages WHERE chat_id=$1 AND is_read=true", chatID).Scan(&readCount)
	s.Require().NoError(err)

	var countA int
	err = s.pool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM messages WHERE chat_id=$1 AND sender_id=$2 AND is_read=true", chatID, userA).Scan(&countA)
	s.Require().NoError(err)
	s.Require().Equal(0, countA, "user A messages should NOT be marked as read")

	var countBUnread int
	err = s.pool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM messages WHERE chat_id=$1 AND sender_id=$2 AND is_read=false", chatID, userB).Scan(&countBUnread)
	s.Require().NoError(err)
	s.Require().Equal(0, countBUnread, "user B messages should all be read (up to lastReadMessageID)")
}

func (s *ChatRepositoryTestSuite) TestMarkMessagesAsRead_NoOtherMessages() {
	userA := s.createUser("user-a2", "pass")

	var chatID int64
	err := s.pool.QueryRow(context.Background(),
		"INSERT INTO chats (name) VALUES('read-chat2') RETURNING id").Scan(&chatID)
	s.Require().NoError(err)

	var msgID int64
	err = s.pool.QueryRow(context.Background(),
		"INSERT INTO messages (chat_id, sender_id, message) VALUES($1,$2,'only-mine') RETURNING id",
		chatID, userA).Scan(&msgID)
	s.Require().NoError(err)

	err = s.repo.MarkMessagesAsRead(context.Background(), chatID, userA, msgID)
	s.Require().NoError(err)

	var readCount int
	err = s.pool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM messages WHERE chat_id=$1 AND is_read=true", chatID).Scan(&readCount)
	s.Require().NoError(err)
	s.Require().Equal(0, readCount, "own messages should not be marked as read")
}
