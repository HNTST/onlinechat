package chat_transport_http

import (
	core_logger "discord/internal/core/logger"
	core_http_request "discord/internal/core/transport/request"
	core_http_response "discord/internal/core/transport/response"
	core_http_utils "discord/internal/core/transport/utils"
	"net/http"
)

type AddUserToChatRequest struct {
	Users_ids []int64 `json:"users_ids" validate:"required"`
}
type AddUserToChatResponse struct {
	Chat_id   int64
	Users_ids []int64
}

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
	chat_id, err := core_http_utils.GetIntQueryParam(r, "chat_id")
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"get query param",
		)
	}
	var request AddUserToChatRequest
	if err := core_http_request.DecodeAndValidateRequest(
		r,
		&request,
	); err != nil {
		responseHandler.ErrorResponse(
			err,
			"error to decode of validate json request",
		)
		return
	}
	err = h.chatService.AddUserToChat(
		ctx,
		int64(*chat_id),
		request.Users_ids,
	)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to add user to chat",
		)
		return
	}
	response := AddUserToChatResponse{
		Chat_id:   int64(*chat_id),
		Users_ids: request.Users_ids,
	}
	responseHandler.JSONResponse(response, http.StatusOK)
}
