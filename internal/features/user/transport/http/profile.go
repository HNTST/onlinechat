package users_transport_http

import (
	core_errors "discord/internal/core/errors"
	core_logger "discord/internal/core/logger"
	core_http_middleware "discord/internal/core/transport/middleware"
	core_http_response "discord/internal/core/transport/response"
	"net/http"
)

// GetProfile returns the current user's profile
//
//	@Summary		Мой профиль
//	@Description	Получение профиля текущего пользователя (user_id из сессии)
//	@Tags			v2 / Пользователи
//	@Produce		json
//	@Success		200	{object}	GetUserResponse
//	@Failure		401	{object}	object{}
//	@Router			/api/v2/user/profile [get]
func (h *UsersHTTPHandler) GetProfile(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	userID, ok := ctx.Value(core_http_middleware.UserIDKey).(int64)
	if !ok {
		responseHandler.ErrorResponse(core_errors.ErrUnauthorized, "не авторизован")
		return
	}

	id := int(userID)
	userDomain, err := h.usersService.GetUser(ctx, &id)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get profile")
		return
	}
	response := GetUserResponse(UserDtoFromDomain(userDomain))

	responseHandler.JSONResponse(response, http.StatusOK)
}
