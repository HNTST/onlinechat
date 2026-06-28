package chat_transport_http

import (
	core_errors "discord/internal/core/errors"
	"discord/internal/core/domain"
	core_logger "discord/internal/core/logger"
	core_http_request "discord/internal/core/transport/request"
	core_http_response "discord/internal/core/transport/response"
	"net/http"
	"time"
)

type SaveMessageRequest struct {
	SenderId int64  `json:"sender_id" validate:"required"`
	Message  string `json:"message" validate:"required"`
}

// SaveMessage creates a new message in a chat
//
//	@Summary		Отправка сообщения
//	@Description	Отправка сообщения в чат
//	@Tags			v2 / Чаты
//	@Accept			json
//	@Produce		json
//	@Param			chat_id	path		int64				true	"ID чата"
//	@Param			request	body		SaveMessageRequest	true	"Данные сообщения"
//	@Success		201		{object}	object{}
//	@Failure		400		{object}	object{}
//	@Router			/api/v2/chats/{chat_id}/messages [post]
func (h *ChatHttpHandler) SaveMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, w)

	chatIDStr := r.PathValue("chat_id")
	if chatIDStr == "" {
		responseHandler.ErrorResponse(core_errors.ErrInvalidArgument, "param is incorrect")
		return
	}
	chatID, err := parseInt64(chatIDStr)
	if err != nil {
		responseHandler.ErrorResponse(core_errors.ErrInvalidArgument, "param is incorrect")
		return
	}

	var req SaveMessageRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &req); err != nil {
		responseHandler.ErrorResponse(core_errors.ErrInvalidArgument, "error to decode or validate json request")
		return
	}

	msg := domain.Messages{
		ChatId:   chatID,
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
