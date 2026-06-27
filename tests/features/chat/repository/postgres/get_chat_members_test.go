package chats_postgres_repository_test

import (
	"context"
)

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
