package chat_transport_http

import (
	core_errors "discord/internal/core/errors"
	core_logger "discord/internal/core/logger"
	core_http_response "discord/internal/core/transport/response"
	"net/http"
	"strconv"
	"time"
)

type GetChatMessagesResponse struct {
	Messages []MessageDTO `json:"messages"`
}
type MessageDTO struct {
	Id        int64     `json:"id"`
	SenderId  int64     `json:"sender_id"`
	Message   string    `json:"message"`
	SentAt    time.Time `json:"sent_at"`
	IsEdited  bool      `json:"is_edited"`
	IsDeleted bool      `json:"is_deleted"`
	IsRead    bool      `json:"is_read"`
}

// GetChatMessages returns messages for a chat with cursor-based pagination
//
//	@Summary		Сообщения чата
//	@Description	Получение сообщений в чате с курсорной пагинацией
//	@Tags			v2 / Чаты
//	@Accept			json
//	@Produce		json
//	@Param			chat_id			query		int	true	"ID чата"
//	@Param			last_message_id	query		int	false	"ID последнего сообщения для пагинации"
//	@Param			limit			query		int	false	"Лимит сообщений (default 50)"
//	@Success		200				{object}	GetChatMessagesResponse
//	@Failure		400				{object}	object{}
//	@Router			/api/v2/messages [get]
func (h *ChatHttpHandler) GetChatMessages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, w)

	chatIDStr := r.URL.Query().Get("chat_id")
	if chatIDStr == "" {
		responseHandler.ErrorResponse(core_errors.ErrParamMissing, "param is incorrect")
		return
	}
	chatID, err := parseInt64(chatIDStr)
	if err != nil {
		responseHandler.ErrorResponse(core_errors.ErrInvalidArgument, "param is incorrect")
		return
	}

	var lastMessageID *int64
	if lmStr := r.URL.Query().Get("last_message_id"); lmStr != "" {
		lm, err := strconv.ParseInt(lmStr, 10, 64)
		if err != nil || lm <= 0 {
			responseHandler.ErrorResponse(core_errors.ErrInvalidArgument, "invalid last_message_id")
			return
		}
		lastMessageID = &lm
	}

	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil || l <= 0 || l > 200 {
			responseHandler.ErrorResponse(core_errors.ErrInvalidArgument, "invalid limit")
			return
		}
		limit = l
	}

	messages, err := h.chatService.GetChatMessagesPaginated(ctx, chatID, lastMessageID, limit)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get messages")
		return
	}

	dtos := make([]MessageDTO, len(messages))
	for i, m := range messages {
		msgText := m.Message
		if m.IsDeleted {
			msgText = "Сообщение удалено"
		}
		dtos[i] = MessageDTO{
			Id:        m.Id,
			SenderId:  m.SenderId,
			Message:   msgText,
			SentAt:    m.Sent_at,
			IsEdited:  m.IsEdited,
			IsDeleted: m.IsDeleted,
			IsRead:    m.IsRead,
		}
	}
	responseHandler.JSONResponse(GetChatMessagesResponse{Messages: dtos}, http.StatusOK)
}
