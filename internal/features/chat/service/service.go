package chat_service

import (
	"context"
	"discord/internal/core/domain"
)

type ChatService struct {
	chatRepository ChatRepository
}

type ChatRepository interface {
	Transaction(
		ctx context.Context,
		fn func(txCtx context.Context) error,
	) error

	CreateNewChat(
		ctx context.Context,
		chat_name string,
	) (int64, error)
	CheckExistsChat(
		ctx context.Context,
		user_id, target_user_id int64,
	) (int64, error)
	AddUserToChat(
		ctx context.Context,
		chat_id int64,
		users_ids []int64,
	) error
	GetUserChats(
		ctx context.Context,
		user_id int64,
	) ([]domain.Chat, error)
	GetChatMessages(
		ctx context.Context,
		chat_id int64,
	) ([]domain.Messages, error)
	GetChatMembers(
		ctx context.Context,
		chat_id int64,
	) ([]int64, error)
	SaveMessage(
		ctx context.Context,
		message domain.Messages,
	) error
}

func NewChatService(chatRepository ChatRepository) *ChatService {
	return &ChatService{
		chatRepository: chatRepository,
	}
}
