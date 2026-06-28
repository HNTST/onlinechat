package chat_transport_http

import (
	core_errors "discord/internal/core/errors"
	core_logger "discord/internal/core/logger"
	core_http_request "discord/internal/core/transport/request"
	core_http_response "discord/internal/core/transport/response"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type EditMessageRequest struct {
	Text string `json:"text" validate:"required"`
}

// EditMessage edits a message text
//
//	@Summary		Редактирование сообщения
//	@Description	Редактирование текста сообщения (устанавливает is_edited=true)
//	@Tags			v2 / Чаты
//	@Accept			json
//	@Produce		json
//	@Param			message_id	path		int					true	"ID сообщения"
//	@Param			request		body		EditMessageRequest	true	"Новый текст"
//	@Success		200			{object}	object{}
//	@Failure		400			{object}	object{}
//	@Router			/api/v2/chat/message/{message_id} [put]
func (h *ChatHttpHandler) EditMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, w)

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

	var req EditMessageRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &req); err != nil {
		responseHandler.ErrorResponse(core_errors.ErrInvalidArgument, "error to decode or validate json request")
		return
	}

	if strings.TrimSpace(req.Text) == "" {
		responseHandler.ErrorResponse(core_errors.ErrInvalidArgument, "text cannot be empty")
		return
	}

	if err := h.chatService.EditMessage(ctx, messageID, req.Text); err != nil {
		if errors.Is(err, core_errors.ErrMessageNotFound) {
			responseHandler.ErrorResponse(core_errors.ErrMessageNotFound, "message not found")
			return
		}
		responseHandler.ErrorResponse(err, "failed to edit message")
		return
	}
	w.WriteHeader(http.StatusOK)
}
