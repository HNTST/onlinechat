package chat_transport_http

import (
	"context"
	"discord/internal/core/domain"
	core_http_server "discord/internal/core/transport/server"
	"net/http"
	"strconv"
)

func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

type ChatHttpHandler struct {
	chatService ChatService
}
type ChatService interface {
	SaveChat(
		ctx context.Context,
		name string,
		user_id,
		target_user_id int64,
	) (int64, error)
	AddUserToChat(
		ctx context.Context,
		chat_id int64,
		user_id []int64,
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
	EditMessage(
		ctx context.Context,
		messageID int64,
		text string,
	) error
	SoftDeleteMessage(
		ctx context.Context,
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
}

func NewChatHttpHandler(
	chatService ChatService,
) *ChatHttpHandler {
	return &ChatHttpHandler{
		chatService: chatService,
	}
}

func (h *ChatHttpHandler) Routes() []core_http_server.Route {
	return []core_http_server.Route{
		{
			Method:  http.MethodPost,
			Path:    "/chat",
			Handler: h.SaveChat,
		},
		{
			Method:  http.MethodPost,
			Path:    "/chat/{chat_id}/members",
			Handler: h.AddUserToChat,
		},
		{
			Method:  http.MethodGet,
			Path:    "/chats",
			Handler: h.GetUserChats,
		},
		{
			Method:  http.MethodGet,
			Path:    "/messages",
			Handler: h.GetChatMessages,
		},
		{
			Method:  http.MethodGet,
			Path:    "/chats/members",
			Handler: h.GetChatMembers,
		},
		{
			Method:  http.MethodPost,
			Path:    "/chat/message",
			Handler: h.SaveMessage,
		},
		{
			Method:  http.MethodPut,
			Path:    "/chat/message/{message_id}",
			Handler: h.EditMessage,
		},
		{
			Method:  http.MethodDelete,
			Path:    "/chat/message/{message_id}",
			Handler: h.DeleteMessage,
		},
		{
			Method:  http.MethodPost,
			Path:    "/chat/messages/read",
			Handler: h.MarkMessagesAsRead,
		},
	}
}
