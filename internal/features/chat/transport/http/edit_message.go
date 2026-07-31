package chat_transport_http

import (
	core_errors "discord/internal/core/errors"
	core_logger "discord/internal/core/logger"
	core_http_middleware "discord/internal/core/transport/middleware"
	core_http_request "discord/internal/core/transport/request"
	core_http_response "discord/internal/core/transport/response"
	"errors"
	"net/http"
	"strings"
)

type EditMessageRequest struct {
	Text string `json:"text" validate:"required"`
}

// EditMessage edits a message text
//
//	@Summary		Редактирование сообщения
//	@Description	Редактирование текста сообщения (устанавливает is_edited=true). Проверяется членство в чате и авторство сообщения.
//	@Tags			v2 / Чаты
//	@Accept			json
//	@Produce		json
//	@Param			chat_id		path		int64				true	"ID чата"
//	@Param			message_id	path		int64				true	"ID сообщения"
//	@Param			request		body		EditMessageRequest	true	"Новый текст"
//	@Success		200			{object}	object{}
//	@Failure		400	{object}	object{}
//	@Failure		404	{object}	object{}
//	@Router			/api/v2/chats/{chat_id}/messages/{message_id} [put]
func (h *ChatHttpHandler) EditMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, w)

	chatIDStr := r.PathValue("chat_id")
	if chatIDStr == "" {
		responseHandler.ErrorResponse(core_errors.ErrInvalidArgument, "missing chat_id in path")
		return
	}
	chatID, err := parseInt64(chatIDStr)
	if err != nil {
		responseHandler.ErrorResponse(core_errors.ErrInvalidArgument, "invalid chat_id in path")
		return
	}

	messageIDStr := r.PathValue("message_id")
	if messageIDStr == "" {
		responseHandler.ErrorResponse(core_errors.ErrInvalidArgument, "missing message_id in path")
		return
	}
	messageID, err := parseInt64(messageIDStr)
	if err != nil {
		responseHandler.ErrorResponse(core_errors.ErrInvalidArgument, "invalid message_id in path")
		return
	}

	var req EditMessageRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &req); err != nil {
		responseHandler.ErrorResponse(core_errors.ErrInvalidArgument, "error to decode or validate json request")
		return
	}

	if strings.TrimSpace(req.Text) == "" {
		responseHandler.ErrorResponse(core_errors.ErrInvalidArgument, "text cannot be empty")
		return
	}

	userID, ok := ctx.Value(core_http_middleware.UserIDKey).(int64)
	if !ok {
		responseHandler.ErrorResponse(core_errors.ErrUnauthorized, "не авторизован")
		return
	}

	if err := h.chatService.EditMessage(ctx, chatID, messageID, userID, req.Text); err != nil {
		if errors.Is(err, core_errors.ErrMessageNotFound) {
			responseHandler.ErrorResponse(core_errors.ErrMessageNotFound, "message not found")
			return
		}
		if errors.Is(err, core_errors.ErrUnauthorized) {
			responseHandler.ErrorResponse(core_errors.ErrUnauthorized, "message not owned by user")
			return
		}
		responseHandler.ErrorResponse(err, "failed to edit message")
		return
	}
	w.WriteHeader(http.StatusOK)
}
