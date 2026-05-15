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
	SaveUser(ctx context.Context, user domain.User) (domain.User, error)
	GetUser(ctx context.Context, user_id *int) (domain.User, error)
}

func NewUsersHTTPHandler(usersService UsersService) *UsersHTTPHandler {
	return &UsersHTTPHandler{
		usersService: usersService,
	}
}
func (h *UsersHTTPHandler) Routes() []core_http_server.Route {
	return []core_http_server.Route{
		{
			Method:  http.MethodPost,
			Path:    "/user",
			Handler: h.SaveUser,
		},
		{
			Method:  http.MethodGet,
			Path:    "/user",
			Handler: h.GetUser,
		},
	}
}
