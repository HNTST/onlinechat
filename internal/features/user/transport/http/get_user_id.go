package users_transport_http

import (
	core_errors "discord/internal/core/errors"
	core_logger "discord/internal/core/logger"
	core_http_response "discord/internal/core/transport/response"
	"errors"
	"net/http"
)

type GetUserIDResponse struct {
	Id int64 `json:"user_id"`
}

func (h *UsersHTTPHandler) GetUserID(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	username := r.URL.Query().Get("username")
	if username == "" {
		responseHandler.ErrorResponse(core_errors.ErrParamMissing, "param is incorrect")
		return
	}

	id, err := h.usersService.GetUserID(ctx, username)
	if err != nil {
		if errors.Is(err, core_errors.ErrUserNotFound) {
			responseHandler.ErrorResponse(core_errors.ErrUserNotFound, "user not found")
			return
		}
		responseHandler.ErrorResponse(err, "failed to get user_id")
		return
	}
	response := GetUserIDResponse{Id: id}
	responseHandler.JSONResponse(response, http.StatusOK)
}
