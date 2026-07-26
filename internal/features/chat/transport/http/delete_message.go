package chat_transport_http

import (
	core_errors "discord/internal/core/errors"
	core_logger "discord/internal/core/logger"
	core_http_response "discord/internal/core/transport/response"
	"errors"
	"net/http"
	"strconv"
)

// DeleteMessage soft-deletes a message
//
//	@Summary		Мягкое удаление сообщения
//	@Description	Помечает сообщение как удалённое (is_deleted=true, текст очищается). Проверяется принадлежность сообщения чату.
//	@Tags			v2 / Чаты
//	@Accept			json
//	@Produce		json
//	@Param			chat_id		path		int64	true	"ID чата"
//	@Param			message_id	path		int64	true	"ID сообщения"
//	@Success		200	{object}	object{}
//	@Failure		400	{object}	object{}
//	@Failure		403	{object}	object{}
//	@Failure		404	{object}	object{}
//	@Router			/api/v2/chats/{chat_id}/messages/{message_id} [delete]
func (h *ChatHttpHandler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, w)

	chatIDStr := r.PathValue("chat_id")
	if chatIDStr == "" {
		responseHandler.ErrorResponse(core_errors.ErrInvalidArgument, "missing chat_id in path")
		return
	}
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		responseHandler.ErrorResponse(core_errors.ErrInvalidArgument, "invalid chat_id in path")
		return
	}

	messageIDStr := r.PathValue("message_id")
	if messageIDStr == "" {
		responseHandler.ErrorResponse(core_errors.ErrInvalidArgument, "missing message_id in path")
		return
	}
	messageID, err := strconv.ParseInt(messageIDStr, 10, 64)
	if err != nil {
		responseHandler.ErrorResponse(core_errors.ErrInvalidArgument, "invalid message_id in path")
		return
	}

	if err := h.chatService.SoftDeleteMessage(ctx, chatID, messageID); err != nil {
		if errors.Is(err, core_errors.ErrMessageNotFound) {
			responseHandler.ErrorResponse(core_errors.ErrMessageNotFound, "message not found")
			return
		}
		if errors.Is(err, core_errors.ErrUnauthorized) {
			responseHandler.ErrorResponse(core_errors.ErrUnauthorized, "message belongs to another chat")
			return
		}
		responseHandler.ErrorResponse(err, "failed to delete message")
		return
	}
	w.WriteHeader(http.StatusOK)
}
