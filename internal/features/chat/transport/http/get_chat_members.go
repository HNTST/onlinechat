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

func (h *ChatHttpHandler) GetChatMembers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, w)

	chatIDStr := r.URL.Query().Get("chat_id")
	if chatIDStr == "" {
		responseHandler.ErrorResponse(core_errors.ErrParamMissing, "param is incorrect")
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
