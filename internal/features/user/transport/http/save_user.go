package users_transport_http

import (
	"discord/internal/core/domain"
	core_logger "discord/internal/core/logger"
	core_http_request "discord/internal/core/transport/request"
	core_http_response "discord/internal/core/transport/response"
	"net/http"
)

type SaveUserRequest struct {
	Username string `json:"username" validate:"required, min=3,max=100"`
	Password string `json:"password" validate:"omitempty,min=8"` // Временно не обязательный
}
type SaveUserResponse UserDTOResponse

// SaveUser creates a new user
//
//	@Summary		Register user
//	@Description	Create a new user account
//	@Tags			v2 / User
//	@Accept			json
//	@Produce		json
//	@Param			request	body		SaveUserRequest	true	"Registration data"
//	@Success		201		{object}	SaveUserResponse
//	@Failure		400		{object}	object{}
//	@Router			/api/v2/user [post]
func (h *UsersHTTPHandler) SaveUser(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)
	var request SaveUserRequest

	if err := core_http_request.DecodeAndValidateRequest(r, &request); err != nil {
		responseHandler.ErrorResponse(err, "ошибка декодирования и валидации HTTP запроса")
		return
	}

	userDomain := domainFromDTO(request)
	userDomain, err := h.usersService.SaveUser(ctx, userDomain)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to save user")
		return
	}
	response := SaveUserResponse(UserDtoFromDomain(userDomain))

	responseHandler.JSONResponse(response, http.StatusCreated)
}

func domainFromDTO(dto SaveUserRequest) domain.User {
	return domain.NewUserUninitialized(dto.Username, dto.Password)
}
