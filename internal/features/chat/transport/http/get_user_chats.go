package chat_transport_http

import (
	core_errors "discord/internal/core/errors"
	core_logger "discord/internal/core/logger"
	core_http_response "discord/internal/core/transport/response"
	"net/http"
	"time"
)

type GetUserChatsResponse struct {
	Chats []ChatDTO `json:"chats"`
}
type ChatDTO struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

type GetUserChatsExtendedResponse struct {
	Chats []ChatExtendedDTO `json:"chats"`
}
type LastMessageDTO struct {
	Id       int64     `json:"id"`
	Text     string    `json:"text"`
	SenderId int64     `json:"sender_id"`
	SentAt   time.Time `json:"sent_at"`
	IsRead   bool      `json:"is_read"`
}
type ChatExtendedDTO struct {
	Id          int64           `json:"id"`
	Name        string          `json:"name"`
	LastMessage *LastMessageDTO `json:"last_message,omitempty"`
	UnreadCount int             `json:"unread_count"`
}

// GetUserChats returns all chats for a user with last message and unread count
//
//	@Summary		Чаты пользователя
//	@Description	Получение всех чатов пользователя с последним сообщением и счётчиком непрочитанных
//	@Tags			v2 / Чаты
//	@Accept			json
//	@Produce		json
//	@Param			user_id	query		int64	true	"ID пользователя"
//	@Success		200		{object}	GetUserChatsExtendedResponse
//	@Failure		400		{object}	object{}
//	@Router			/api/v2/chats [get]
func (h *ChatHttpHandler) GetUserChats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, w)

	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		responseHandler.ErrorResponse(core_errors.ErrInvalidArgument, "param is incorrect")
		return
	}
	userID, err := parseInt64(userIDStr)
	if err != nil {
		responseHandler.ErrorResponse(core_errors.ErrInvalidArgument, "param is incorrect")
		return
	}

	chats, err := h.chatService.GetUserChatsExtended(ctx, userID)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get user's chats")
		return
	}
	dtos := make([]ChatExtendedDTO, len(chats))
	for i, c := range chats {
		var lm *LastMessageDTO
		if c.LastMessage != nil {
			lm = &LastMessageDTO{
				Id:       c.LastMessage.Id,
				Text:     c.LastMessage.Text,
				SenderId: c.LastMessage.SenderId,
				SentAt:   c.LastMessage.SentAt,
				IsRead:   c.LastMessage.IsRead,
			}
		}
		dtos[i] = ChatExtendedDTO{
			Id:          c.Id,
			Name:        c.Name,
			LastMessage: lm,
			UnreadCount: c.UnreadCount,
		}
	}
	responseHandler.JSONResponse(GetUserChatsExtendedResponse{Chats: dtos}, http.StatusOK)
}
