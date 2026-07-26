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
		chatName string,
		isGroup bool,
	) (int64, error)
	CheckExistsChat(
		ctx context.Context,
		userID, targetUserID int64,
	) (int64, error)
	AddUserToChat(
		ctx context.Context,
		chatID int64,
		userIDs []int64,
	) error
	GetUserChats(
		ctx context.Context,
		userID int64,
	) ([]domain.Chat, error)
	GetChatMessages(
		ctx context.Context,
		chatID int64,
	) ([]domain.Messages, error)
	GetChatMembers(
		ctx context.Context,
		chatID int64,
	) ([]int64, error)
	SaveMessage(
		ctx context.Context,
		message domain.Messages,
	) (int64, error)

	EditMessage(
		ctx context.Context,
		messageID int64,
		text string,
	) error
	SoftDeleteMessage(
		ctx context.Context,
		chatID int64,
		messageID int64,
	) error
	MarkMessagesAsRead(
		ctx context.Context,
		chatID int64,
		userID int64,
		lastReadMessageID int64,
	) error
	GetChatMessagesPaginated(
		ctx context.Context,
		chatID int64,
		lastMessageID *int64,
		limit int,
	) ([]domain.Messages, error)
	GetUserChatsExtended(
		ctx context.Context,
		userID int64,
	) ([]domain.ChatInfo, error)

	UpdateChatName(
		ctx context.Context,
		chatID int64,
		newName string,
	) error
	DeleteChat(
		ctx context.Context,
		chatID int64,
	) error
	GetChatPartnerIDs(
		ctx context.Context,
		userID int64,
	) ([]int64, error)
	GetUserChatIDs(
		ctx context.Context,
		userID int64,
	) ([]int64, error)
	GetChatByID(
		ctx context.Context,
		userID int64,
		chatID int64,
	) (domain.Chat, error)
}

func NewChatService(chatRepository ChatRepository) *ChatService {
	return &ChatService{
		chatRepository: chatRepository,
	}
}
