package chat_transport_http

import (
	core_errors "discord/internal/core/errors"
	core_logger "discord/internal/core/logger"
	core_http_response "discord/internal/core/transport/response"
	core_http_utils "discord/internal/core/transport/utils"
	"net/http"
)

// DeleteChat deletes a group chat
//
//	@Summary		Удаление группового чата
//	@Description	Удаление группового чата. Только для групповых чатов (is_group=true).
//	@Tags			v2 / Чаты
//	@Produce		json
//	@Param			chat_id	path		int64	true	"ID чата"
//	@Param			user_id	query		int64	true	"ID пользователя"
//	@Success		200		{object}	object{}
//	@Failure		400		{object}	object{}
//	@Router			/api/v2/chat/{chat_id} [delete]
func (h *ChatHttpHandler) DeleteChat(
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

	userIDParam, err := core_http_utils.GetIntQueryParam(r, "user_id")
	if err != nil {
		responseHandler.ErrorResponse(core_errors.ErrInvalidArgument, "invalid user_id")
		return
	}
	if userIDParam == nil {
		responseHandler.ErrorResponse(core_errors.ErrInvalidArgument, "missing user_id")
		return
	}
	userID := int64(*userIDParam)

	err = h.chatService.DeleteChat(ctx, chatID, userID)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to delete chat")
		return
	}

	responseHandler.JSONResponse(map[string]string{}, http.StatusOK)
}
