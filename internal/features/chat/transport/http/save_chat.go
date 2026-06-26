package chat_transport_http

import (
	core_logger "discord/internal/core/logger"
	core_http_request "discord/internal/core/transport/request"
	core_http_response "discord/internal/core/transport/response"
	"net/http"
)

type SaveChatRequest struct {
	ChatName   string `json:"chat_name" validate:"required"`
	Author     int64  `json:"author" validate:"required"`
	TargetUser int64  `json:"target_user" validate:"required"`
}
type SaveChatResponse struct {
	Id int64 `json:"chat_id"`
}

// SaveChat creates a new chat
//
//	@Summary		Create chat
//	@Description	Create a new chat with two participants
//	@Tags			v2 / Chat
//	@Accept			json
//	@Produce		json
//	@Param			request	body		SaveChatRequest	true	"Chat data"
//	@Success		201		{object}	SaveChatResponse
//	@Failure		400		{object}	object{}
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
			err,
			"error to decode or validate json request",
		)
		return
	}
	chat_id, err := h.chatService.SaveChat(
		ctx,
		request.ChatName,
		request.Author,
		request.TargetUser,
	)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to save chat and save members",
		)
		return
	}

	response := SaveChatResponse{Id: chat_id}

	responseHandler.JSONResponse(response, http.StatusCreated)
}
