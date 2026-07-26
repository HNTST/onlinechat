package chat_transport_http

import (
	core_errors "discord/internal/core/errors"
	core_logger "discord/internal/core/logger"
	core_http_request "discord/internal/core/transport/request"
	core_http_response "discord/internal/core/transport/response"
	"net/http"
	"strconv"
)

type AddUserToChatRequest struct {
	UserIDs []int64 `json:"users_ids" validate:"required"`
}
type AddUserToChatResponse struct {
	ChatID  int64   `json:"Chat_id"`
	UserIDs []int64 `json:"Users_ids"`
}

// AddUserToChat adds users to an existing chat
//
//	@Summary		Добавление пользователей в чат
//	@Description	Добавление пользователей в существующий чат по его ID
//	@Tags			v2 / Чаты
//	@Accept			json
//	@Produce		json
//	@Param			chat_id	path		int64				true	"ID чата"
//	@Param			request	body		AddUserToChatRequest	true	"ID пользователей"
//	@Success		200		{object}	AddUserToChatResponse
//	@Failure		400		{object}	object{}
//	@Router			/api/v2/chat/{chat_id}/members [post]
func (h *ChatHttpHandler) AddUserToChat(
	w http.ResponseWriter,
	r *http.Request,
) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(
		log,
		w,
	)
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
	var request AddUserToChatRequest
	if err := core_http_request.DecodeAndValidateRequest(
		r,
		&request,
	); err != nil {
		responseHandler.ErrorResponse(
			core_errors.ErrInvalidArgument,
			"error to decode of validate json request",
		)
		return
	}
	err = h.chatService.AddUserToChat(
		ctx,
		chatID,
		request.UserIDs,
	)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to add user to chat",
		)
		return
	}
	response := AddUserToChatResponse{
		ChatID:  chatID,
		UserIDs: request.UserIDs,
	}
	responseHandler.JSONResponse(response, http.StatusOK)
}
