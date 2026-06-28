package chats_postgres_repository_test

import (
	"context"
)

func (s *ChatRepositoryTestSuite) TestGetChatMessagesPaginated_NoCursor() {
	userID := s.createUser("paginated-user", "pass")

	var chatID int64
	err := s.pool.QueryRow(context.Background(),
		"INSERT INTO chats (name) VALUES('paginated-chat') RETURNING id").Scan(&chatID)
	s.Require().NoError(err)

	for i := 0; i < 10; i++ {
		_, err = s.pool.Exec(context.Background(),
			"INSERT INTO messages (chat_id, sender_id, message) VALUES($1,$2,$3)",
			chatID, userID, "msg-"+string(rune('0'+i)))
		s.Require().NoError(err)
	}

	msgs, err := s.repo.GetChatMessagesPaginated(context.Background(), chatID, nil, 5)
	s.Require().NoError(err)
	s.Require().Len(msgs, 5)
	s.Require().Equal("msg-5", msgs[0].Message, "should return the 5 most recent messages")
}

func (s *ChatRepositoryTestSuite) TestGetChatMessagesPaginated_WithCursor() {
	userID := s.createUser("cursor-user", "pass")

	var chatID int64
	err := s.pool.QueryRow(context.Background(),
		"INSERT INTO chats (name) VALUES('cursor-chat') RETURNING id").Scan(&chatID)
	s.Require().NoError(err)

	for i := 0; i < 10; i++ {
		_, err = s.pool.Exec(context.Background(),
			"INSERT INTO messages (chat_id, sender_id, message) VALUES($1,$2,$3)",
			chatID, userID, "msg-"+string(rune('0'+i)))
		s.Require().NoError(err)
	}

	var fifthMsgID int64
	err = s.pool.QueryRow(context.Background(),
		"SELECT id FROM messages WHERE chat_id=$1 ORDER BY id ASC LIMIT 1 OFFSET 4", chatID).Scan(&fifthMsgID)
	s.Require().NoError(err)

	msgs, err := s.repo.GetChatMessagesPaginated(context.Background(), chatID, &fifthMsgID, 3)
	s.Require().NoError(err)
	s.Require().Len(msgs, 3)
	s.Require().Equal("msg-1", msgs[0].Message, "should return 3 messages before msg-4")
	s.Require().Equal("msg-2", msgs[1].Message)
	s.Require().Equal("msg-3", msgs[2].Message)
}

func (s *ChatRepositoryTestSuite) TestGetChatMessagesPaginated_Empty() {
	var chatID int64
	err := s.pool.QueryRow(context.Background(),
		"INSERT INTO chats (name) VALUES('empty-paginated') RETURNING id").Scan(&chatID)
	s.Require().NoError(err)

	msgs, err := s.repo.GetChatMessagesPaginated(context.Background(), chatID, nil, 50)
	s.Require().NoError(err)
	s.Require().Empty(msgs)
}
