package domain

import (
	"context"
	"io"
)

type Repository interface {
	InitDB() error

	GetUser(id int64) (User, error)
	GetUserID(username string) (int64, error)
	GetUserChats(user_id int64) ([]Chat, error)
	GetChatMessages(chat_id int64) ([]Messages, error)
	GetChatMembers(chat_id int64) ([]int64, error)

	SaveUser(username, password string) (int64, error)
	SaveChatWithAuthor(name string, user_id, target_user_id int64) (int64, error)
	SaveMessage(message Messages) error
	SaveUserToChat(user_id, chat_id int64) error

	UpdateUserAvatarPath(user_id, size int64, file_key, orig_file_name, type_file string) (int64, error)
}

type Service interface {
	SaveUser(username, password string) (int64, error)
	SaveChat(chat_name string, user_id, target_user_id int64) (int64, error)
	GetUser(user_id int64) (User, error)
	GetUserChats(user_id int64) ([]Chat, error)
	GetUserMessages(chat_id int64) ([]Messages, error)
	GetUserID(username string) (int64, error)
	GetChatMembers(chat_id int64) ([]int64, error)
	SaveMessage(message Messages) error

	UploadUserAvatar(ctx context.Context, user_id int64, file io.Reader, orig_file_name string, size int64, content_type string) (int64, error)
}

type Storage interface {
	UploadFile(ctx context.Context, file_name, content_type string, reader io.Reader, size int64) (string, error)
}
