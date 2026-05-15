package chat_transport_http

import (
	"context"
	core_http_server "discord/internal/core/transport/server"
	"net/http"
)

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
	}
}
