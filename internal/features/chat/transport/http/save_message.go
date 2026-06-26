package chat_transport_http

import (
	"discord/internal/core/domain"
	core_logger "discord/internal/core/logger"
	core_http_request "discord/internal/core/transport/request"
	core_http_response "discord/internal/core/transport/response"
	"net/http"
	"time"
)

type SaveMessageRequest struct {
	ChatId   int64  `json:"chat_id" validate:"required"`
	SenderId int64  `json:"sender_id" validate:"required"`
	Message  string `json:"message" validate:"required"`
}

// SaveMessage creates a new message in a chat
//
//	@Summary		Send message
//	@Description	Send a message to a chat
//	@Tags			v2 / Chat
//	@Accept			json
//	@Produce		json
//	@Param			request	body		SaveMessageRequest	true	"Message data"
//	@Success		201		{object}	object{}
//	@Failure		400		{object}	object{}
//	@Router			/api/v2/chat/message [post]
func (h *ChatHttpHandler) SaveMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, w)

	var req SaveMessageRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &req); err != nil {
		responseHandler.ErrorResponse(err, "error to decode or validate json request")
		return
	}

	msg := domain.Messages{
		ChatId:   req.ChatId,
		SenderId: req.SenderId,
		Message:  req.Message,
		Sent_at:  time.Now(),
	}
	if err := h.chatService.SaveMessage(ctx, msg); err != nil {
		responseHandler.ErrorResponse(err, "failed to save message")
		return
	}
	w.WriteHeader(http.StatusCreated)
}
