package chat_transport_http

import (
	core_errors "discord/internal/core/errors"
	core_logger "discord/internal/core/logger"
	core_http_response "discord/internal/core/transport/response"
	"net/http"
)

type GetChatMembersResponse struct {
	Members []int64 `json:"members"`
}

// GetChatMembers returns member IDs for a chat
//
//	@Summary		Участники чата
//	@Description	Получение ID всех участников чата
//	@Tags			v2 / Чаты
//	@Accept			json
//	@Produce		json
//	@Param			chat_id	path		int	true	"ID чата"
//	@Success		200		{object}	GetChatMembersResponse
//	@Failure		400		{object}	object{}
//	@Router			/api/v2/chats/{chat_id}/members [get]
func (h *ChatHttpHandler) GetChatMembers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, w)

	chatIDStr := r.PathValue("chat_id")
	if chatIDStr == "" {
		responseHandler.ErrorResponse(core_errors.ErrInvalidArgument, "param is incorrect")
		return
	}
	chatID, err := parseInt64(chatIDStr)
	if err != nil {
		responseHandler.ErrorResponse(core_errors.ErrInvalidArgument, "param is incorrect")
		return
	}

	members, err := h.chatService.GetChatMembers(ctx, chatID)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get chat members")
		return
	}
	responseHandler.JSONResponse(GetChatMembersResponse{Members: members}, http.StatusOK)
}
