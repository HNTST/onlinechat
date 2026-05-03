package service

import (
	"context"
	"discord/internal/core/domain"
	"fmt"
	"io"
)

func (sv *Service) GetUserID(ctx context.Context, username string) (int64, error) {
	return sv.db.GetUserID(ctx, username)
}

func (sv *Service) SaveUser(ctx context.Context, username, password string) (int64, error) {
	return sv.db.SaveUser(ctx, username, password)
}
func (sv *Service) SaveChat(ctx context.Context, chat_name string, user_id, target_user_id int64) (int64, error) {
	return sv.db.SaveChatWithAuthor(ctx, chat_name, user_id, target_user_id)
}

func (sv *Service) GetUser(ctx context.Context, user_id int64) (domain.User, error) {
	return sv.db.GetUser(ctx, user_id)
}
func (sv *Service) GetUserChats(ctx context.Context, user_id int64) ([]domain.Chat, error) {
	return sv.db.GetUserChats(ctx, user_id)
}
func (sv *Service) GetUserMessages(ctx context.Context, chat_id int64) ([]domain.Messages, error) {
	return sv.db.GetChatMessages(ctx, chat_id)
}
func (sv *Service) GetChatMembers(ctx context.Context, chat_id int64) ([]int64, error) {
	return sv.db.GetChatMembers(ctx, chat_id)
}
func (sv *Service) SaveMessage(ctx context.Context, message domain.Messages) error {
	return sv.db.SaveMessage(ctx, message)
}

func (sv *Service) UploadUserAvatar(ctx context.Context, user_id int64, file io.Reader, orig_file_name string, size int64, content_type string) (int64, error) {

	file_name := fmt.Sprintf("avatars/user_%d", user_id)

	key, err := sv.storage.UploadFile(ctx, file_name, content_type, file, size)
	if err != nil {
		return 0, fmt.Errorf("failed to upload to minio: %w", err)
	}
	file_id, err := sv.db.UpdateUserAvatarPath(ctx, user_id, size, key, orig_file_name, content_type)
	if err != nil {
		return 0, fmt.Errorf("failed to update avatar path in db: %w", err)
	}
	return file_id, nil
}
func (sv *Service) AddUserToChat(ctx context.Context, chat_id, target_user_id int64) error {
	return sv.db.SaveUserToChat(ctx, target_user_id, chat_id)
}
