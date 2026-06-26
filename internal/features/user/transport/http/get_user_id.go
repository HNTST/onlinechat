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

// GetUserID returns a user ID by username
//
//	@Summary		Get user ID
//	@Description	Get user ID by username
//	@Tags			v2 / User
//	@Accept			json
//	@Produce		json
//	@Param			username	query		string	true	"Username"
//	@Success		200			{object}	GetUserIDResponse
//	@Failure		400			{object}	object{}
//	@Router			/api/v2/user_id [get]
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
