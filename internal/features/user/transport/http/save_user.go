package users_transport_http

import (
	"discord/internal/core/domain"
	core_logger "discord/internal/core/logger"
	core_http_request "discord/internal/core/transport/request"
	core_http_response "discord/internal/core/transport/response"
	"net/http"
	"time"
)

type SaveUserRequest struct {
	Username string `json:"username" validate:"required, min=3,max=100"`
	Password string `json:"password" validate:"omitempty,min=8"` // Временно не обязательный
}
type SaveUserResponse struct {
	Id   int64     `json:"id"`
	Name string    `json:"username"`
	Reg  time.Time `json:"reg"`
}

func (h *UsersHTTPHandler) SaveUser() (rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)
	var request SaveUserRequest

	if err := core_http_request.DecodeAndValidateRequest(r, &request); err != nil {
		responseHandler.ErrorResponse(err, "ошибка декодирования и валидации HTTP запроса")
		return
	}

	userDomain := domainFromDTO(request)
	userDomain, err := h.usersService.CreateUser(ctx, userDomain)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to create user")
		return
	}
	response := dtoFromDomain(userDomain)

	responseHandler.JSONResponse(response, http.StatusCreated)
	return
}

func domainFromDTO(dto SaveUserRequest) domain.User {
	return domain.NewUserUninitialized(dto.Username, dto.Password)
}

func dtoFromDomain(user domain.User) SaveUserResponse {
	return SaveUserResponse{
		Id:   user.Id,
		Name: user.Name,
		Reg:  user.Reg,
	}
}
