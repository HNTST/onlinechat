package service

import (
	"context"
	"discord/internal/core/domain"
	"io"
)

type Service struct {
	db      Repository
	storage Storage
}

type Repository interface {
	GetUserID(ctx context.Context, username string) (int64, error)
	GetUser(ctx context.Context, id int64) (domain.User, error)
	FindExistingChat(ctx context.Context, user1, user2 int64) (int64, error)
	GetUserChats(ctx context.Context, user_id int64) ([]domain.Chat, error)
	GetChatMessages(ctx context.Context, chat_id int64) ([]domain.Messages, error)
	GetChatMembers(ctx context.Context, chat_id int64) ([]int64, error)
	SaveUser(ctx context.Context, username, password string) (int64, error)
	SaveChatWithAuthor(ctx context.Context, name string, user_id, target_user_id int64) (int64, error)
	SaveMessage(ctx context.Context, message domain.Messages) error
	SaveUserToChat(ctx context.Context, user_id, chat_id int64) error
	UpdateUserAvatarPath(ctx context.Context, user_id, size int64, file_key, orig_file_name, type_file string) (int64, error)
}
type Storage interface {
	UploadFile(ctx context.Context, file_name, content_type string, reader io.Reader, size int64) (string, error)
}

func NewService(db Repository, storage Storage) *Service {
	return &Service{
		db:      db,
		storage: storage,
	}
}
