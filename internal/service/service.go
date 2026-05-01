package service

import (
	"context"
	"discord/internal/core/domain"
	"fmt"
	"io"
)

type Service struct {
	db      domain.Repository
	storage domain.Storage
}

func NewService(db domain.Repository, storage domain.Storage) *Service {
	return &Service{
		db:      db,
		storage: storage,
	}
}

func (sv *Service) SaveUser(username, password string) (int64, error) {
	return sv.db.SaveUser(username, password)
}
func (sv *Service) SaveChat(chat_name string, user_id, target_user_id int64) (int64, error) {
	return sv.db.SaveChatWithAuthor(chat_name, user_id, target_user_id)
}

func (sv *Service) GetUser(user_id int64) (domain.User, error) {
	return sv.db.GetUser(user_id)
}
func (sv *Service) GetUserChats(user_id int64) ([]domain.Chat, error) {
	return sv.db.GetUserChats(user_id)
}
func (sv *Service) GetUserMessages(chat_id int64) ([]domain.Messages, error) {
	return sv.db.GetChatMessages(chat_id)
}
func (sv *Service) GetUserID(username string) (int64, error) {
	return sv.db.GetUserID(username)
}
func (sv *Service) GetChatMembers(chat_id int64) ([]int64, error) {
	return sv.db.GetChatMembers(chat_id)
}
func (sv *Service) SaveMessage(message domain.Messages) error {
	return sv.db.SaveMessage(message)
}

func (sv *Service) UploadUserAvatar(ctx context.Context, user_id int64, file io.Reader, orig_file_name string, size int64, content_type string) (int64, error) {

	file_name := fmt.Sprintf("avatars/user_%d", user_id)

	key, err := sv.storage.UploadFile(ctx, file_name, content_type, file, size)
	if err != nil {
		return 0, fmt.Errorf("failed to upload to minio: %w", err)
	}
	file_id, err := sv.db.UpdateUserAvatarPath(user_id, size, key, orig_file_name, content_type)
	if err != nil {
		return 0, fmt.Errorf("failed to update avatar path in db: %w", err)
	}
	return file_id, nil
}
