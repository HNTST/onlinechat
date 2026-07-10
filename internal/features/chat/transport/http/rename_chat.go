package chat_transport_http

import (
	core_errors "discord/internal/core/errors"
	core_logger "discord/internal/core/logger"
	core_http_middleware "discord/internal/core/transport/middleware"
	core_http_request "discord/internal/core/transport/request"
	core_http_response "discord/internal/core/transport/response"
	"net/http"
)

type RenameChatRequest struct {
	Name string `json:"name" validate:"required"`
}

// RenameChat renames a group chat
//
//	@Summary		Переименование группового чата
//	@Description	Изменение названия группового чата. Только для групповых чатов (is_group=true).
//	@Tags			v2 / Чаты
//	@Accept			json
//	@Produce		json
//	@Param			chat_id	path		int64				true	"ID чата"
//	@Param			request	body		RenameChatRequest	true	"Новое название"
//	@Success		200		{object}	object{}
//	@Failure		400		{object}	object{}
//	@Router			/api/v2/chat/{chat_id}/name [put]
func (h *ChatHttpHandler) RenameChat(
	w http.ResponseWriter,
	r *http.Request,
) {
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

	userID, ok := ctx.Value(core_http_middleware.UserIDKey).(int64)
	if !ok {
		responseHandler.ErrorResponse(core_errors.ErrUnauthorized, "не авторизован")
		return
	}

	var request RenameChatRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &request); err != nil {
		responseHandler.ErrorResponse(core_errors.ErrInvalidArgument, "error to decode or validate json request")
		return
	}

	err = h.chatService.RenameChat(ctx, chatID, userID, request.Name)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to rename chat")
		return
	}

	responseHandler.JSONResponse(map[string]string{}, http.StatusOK)
}
