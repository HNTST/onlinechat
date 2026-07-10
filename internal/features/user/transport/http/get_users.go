package users_transport_http

import (
	core_errors "discord/internal/core/errors"
	core_logger "discord/internal/core/logger"
	core_http_response "discord/internal/core/transport/response"
	core_http_utils "discord/internal/core/transport/utils"
	"net/http"
)

// GetUsers returns a user by ID (for viewing other profiles)
//
//	@Summary		Просмотр профиля
//	@Description	Получение публичных данных пользователя по его ID
//	@Tags			v2 / Пользователи
//	@Produce		json
//	@Param			id	query		int64	true	"ID пользователя"
//	@Success		200	{object}	GetUserResponse
//	@Failure		400	{object}	object{}
//	@Router			/api/v2/users [get]
func (h *UsersHTTPHandler) GetUsers(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	userIDParam, err := core_http_utils.GetIntQueryParam(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get 'id' query param")
		return
	}
	if userIDParam == nil {
		responseHandler.ErrorResponse(core_errors.ErrInvalidArgument, "missing 'id' query param")
		return
	}

	userDomain, err := h.usersService.GetUser(ctx, userIDParam)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get user")
		return
	}
	response := GetUserResponse(UserDtoFromDomain(userDomain))

	responseHandler.JSONResponse(response, http.StatusOK)
}
