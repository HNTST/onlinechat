package chat_transport_http

import (
	core_errors "discord/internal/core/errors"
	core_logger "discord/internal/core/logger"
	core_http_request "discord/internal/core/transport/request"
	core_http_response "discord/internal/core/transport/response"
	"net/http"
)

type MarkMessagesAsReadRequest struct {
	ChatID           int64 `json:"chat_id" validate:"required"`
	LastReadMessageID int64 `json:"last_read_message_id" validate:"required"`
}

// MarkMessagesAsRead marks messages as read for a user in a chat
//
//	@Summary		Отметка о прочтении
//	@Description	Помечает все чужие сообщения в чате до указанного ID как прочитанные
//	@Tags			v2 / Чаты
//	@Accept			json
//	@Produce		json
//	@Param			request	body		MarkMessagesAsReadRequest	true	"Данные для отметки"
//	@Success		200		{object}	object{}
//	@Failure		400		{object}	object{}
//	@Router			/api/v2/chat/messages/read [post]
func (h *ChatHttpHandler) MarkMessagesAsRead(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, w)

	var req MarkMessagesAsReadRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &req); err != nil {
		responseHandler.ErrorResponse(core_errors.ErrInvalidArgument, "error to decode or validate json request")
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		responseHandler.ErrorResponse(core_errors.ErrInvalidArgument, "user_id query param required")
		return
	}
	userID, err := parseInt64(userIDStr)
	if err != nil {
		responseHandler.ErrorResponse(core_errors.ErrInvalidArgument, "invalid user_id")
		return
	}

	if err := h.chatService.MarkMessagesAsRead(ctx, req.ChatID, userID, req.LastReadMessageID); err != nil {
		responseHandler.ErrorResponse(err, "failed to mark messages as read")
		return
	}
	w.WriteHeader(http.StatusOK)
}
