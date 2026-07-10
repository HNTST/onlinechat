package users_transport_http

import (
	"context"
	"discord/internal/core/domain"
	core_http_server "discord/internal/core/transport/server"
	"net/http"
)

type UsersHTTPHandler struct {
	usersService UsersService
}
type UsersService interface {
	GetUser(ctx context.Context, user_id *int) (domain.User, error)
	GetUserID(ctx context.Context, username string) (int64, error)
}

func NewUsersHTTPHandler(usersService UsersService) *UsersHTTPHandler {
	return &UsersHTTPHandler{
		usersService: usersService,
	}
}
func (h *UsersHTTPHandler) Routes() []core_http_server.Route {
	return []core_http_server.Route{
		{
			Method:  http.MethodGet,
			Path:    "/user",
			Handler: h.GetUser,
		},
		{
			Method:  http.MethodGet,
			Path:    "/user_id",
			Handler: h.GetUserID,
		},
	}
}
