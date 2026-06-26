package chat_transport_http

import (
	core_errors "discord/internal/core/errors"
	core_logger "discord/internal/core/logger"
	core_http_response "discord/internal/core/transport/response"
	"net/http"
	"time"
)

type GetChatMessagesResponse struct {
	Messages []MessageDTO `json:"messages"`
}
type MessageDTO struct {
	SenderId int64     `json:"sender_id"`
	Message  string    `json:"message"`
	SentAt   time.Time `json:"sent_at"`
}

// GetChatMessages returns messages for a chat
//
//	@Summary		Get chat messages
//	@Description	Get all messages in a chat
//	@Tags			v2 / Chat
//	@Accept			json
//	@Produce		json
//	@Param			chat_id	query		int	true	"Chat ID"
//	@Success		200		{object}	GetChatMessagesResponse
//	@Failure		400		{object}	object{}
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

	messages, err := h.chatService.GetChatMessages(ctx, chatID)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get messages")
		return
	}
	dtos := make([]MessageDTO, len(messages))
	for i, m := range messages {
		dtos[i] = MessageDTO{SenderId: m.SenderId, Message: m.Message, SentAt: m.Sent_at}
	}
	responseHandler.JSONResponse(GetChatMessagesResponse{Messages: dtos}, http.StatusOK)
}
