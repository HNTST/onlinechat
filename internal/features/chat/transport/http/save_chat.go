package chat_transport_http

import (
	core_errors "discord/internal/core/errors"
	core_logger "discord/internal/core/logger"
	core_http_middleware "discord/internal/core/transport/middleware"
	core_http_request "discord/internal/core/transport/request"
	core_http_response "discord/internal/core/transport/response"
	"net/http"
)

type SaveChatRequest struct {
	ChatName   string `json:"chat_name" validate:"required"`
	TargetUser int64  `json:"target_user" validate:"required"`
	IsGroup    bool   `json:"is_group"`
}
type SaveChatResponse struct {
	Id int64 `json:"chat_id"`
}

// SaveChat creates a new chat
//
//	@Summary		Создание чата
//	@Description	Создание нового чата с двумя участниками. Автор чата определяется по сессии (не передаётся в теле запроса).
//	@Tags			v2 / Чаты
//	@Accept			json
//	@Produce		json
//	@Param			request	body		SaveChatRequest	true	"Данные чата"
//	@Success		201		{object}	SaveChatResponse
//	@Failure		400		{object}	object{}
//	@Failure		401		{object}	object{}
//	@Router			/api/v2/chat [post]
func (h *ChatHttpHandler) SaveChat(
	w http.ResponseWriter,
	r *http.Request,
) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(
		log,
		w,
	)
	var request SaveChatRequest

	if err := core_http_request.DecodeAndValidateRequest(
		r,
		&request,
	); err != nil {
		responseHandler.ErrorResponse(
			core_errors.ErrInvalidArgument,
			"error to decode or validate json request",
		)
		return
	}
	userID, ok := ctx.Value(core_http_middleware.UserIDKey).(int64)
	if !ok {
		responseHandler.ErrorResponse(core_errors.ErrUnauthorized, "не авторизован")
		return
	}
	chatID, err := h.chatService.SaveChat(
		ctx,
		request.ChatName,
		userID,
		request.TargetUser,
		request.IsGroup,
	)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to save chat and save members",
		)
		return
	}

	response := SaveChatResponse{Id: chatID}

	responseHandler.JSONResponse(response, http.StatusCreated)
}
