package chat_transport_http

import (
	core_errors "discord/internal/core/errors"
	core_logger "discord/internal/core/logger"
	core_http_response "discord/internal/core/transport/response"
	"net/http"
)

type GetUserChatsResponse struct {
	Chats []ChatDTO `json:"chats"`
}
type ChatDTO struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

// GetUserChats returns all chats for a user
//
//	@Summary		Get user chats
//	@Description	Get all chats that a user is a member of
//	@Tags			v2 / Chat
//	@Accept			json
//	@Produce		json
//	@Param			user_id	query		int	true	"User ID"
//	@Success		200		{object}	GetUserChatsResponse
//	@Failure		400		{object}	object{}
//	@Router			/api/v2/chats [get]
func (h *ChatHttpHandler) GetUserChats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, w)

	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		responseHandler.ErrorResponse(core_errors.ErrParamMissing, "param is incorrect")
		return
	}
	userID, err := parseInt64(userIDStr)
	if err != nil {
		responseHandler.ErrorResponse(core_errors.ErrInvalidArgument, "param is incorrect")
		return
	}

	chats, err := h.chatService.GetUserChats(ctx, userID)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get user's chats")
		return
	}
	dtos := make([]ChatDTO, len(chats))
	for i, c := range chats {
		dtos[i] = ChatDTO{Id: c.Id, Name: c.Name}
	}
	responseHandler.JSONResponse(GetUserChatsResponse{Chats: dtos}, http.StatusOK)
}
