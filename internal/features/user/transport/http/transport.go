package users_transport_http

import (
	"context"
	"discord/internal/core/domain"
)

type UsersHTTPHandler struct {
	usersService UsersService
}
type UsersService interface {
	CreateUser(ctx context.Context, user domain.User) (domain.User, error)
}

func NewUsersHTTPHandler(usersService UsersService) *UsersHTTPHandler {
	return &UsersHTTPHandler{
		usersService: usersService,
	}
}
