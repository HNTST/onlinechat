package users_transport_http

import (
	core_logger "discord/internal/core/logger"
	core_http_response "discord/internal/core/transport/response"
	core_http_utils "discord/internal/core/transport/utils"
	"fmt"
	"net/http"
)

type GetUserResponse UserDTOResponse

// GetUser returns a user by ID
//
//	@Summary		Получение пользователя
//	@Description	Получение данных пользователя по его ID
//	@Tags			v2 / Пользователи
//	@Accept			json
//	@Produce		json
//	@Param			user_id	query		int	true	"ID пользователя"
//	@Success		200		{object}	GetUserResponse
//	@Failure		400		{object}	object{}
//	@Router			/api/v2/user [get]
func (h *UsersHTTPHandler) GetUser(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)
	user_id, err := getUserIdQueryParams(r)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to get 'user' query param",
		)
		return
	}
	userDomain, err := h.usersService.GetUser(ctx, user_id)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to get users",
		)
		return
	}
	response := GetUserResponse(UserDtoFromDomain(userDomain))

	responseHandler.JSONResponse(response, http.StatusOK)

}
func getUserIdQueryParams(r *http.Request) (*int, error) {
	user_id, err := core_http_utils.GetIntQueryParam(r, "user_id")
	if err != nil {
		return nil, fmt.Errorf("get 'user_id' query param: %w", err)
	}
	return user_id, nil
}
