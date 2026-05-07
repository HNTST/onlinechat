package chat_service

import "context"

type ChatService struct {
	chatRepository ChatRepository
}

type ChatRepository interface {
	SaveChat(
		ctx context.Context,
		name string,
		user_id, target_user_id int64,
	) (int64, error)
}

func NewChatService(chatRepository ChatRepository) *ChatService {
	return &ChatService{
		chatRepository: chatRepository,
	}
}
