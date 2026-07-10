package auth_transport_http

import (
	"context"
	"discord/internal/core/domain"
	core_http_server "discord/internal/core/transport/server"
	"net/http"
)

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthHTTPHandler struct {
	authService AuthService
}

type AuthService interface {
	Login(ctx context.Context, username, password string) (token string, user domain.User, err error)
	Logout(ctx context.Context, token string) error
	ValidateSession(ctx context.Context, token string) (int64, error)
}

func NewAuthHTTPHandler(authService AuthService) *AuthHTTPHandler {
	return &AuthHTTPHandler{
		authService: authService,
	}
}

func (h *AuthHTTPHandler) Routes() []core_http_server.Route {
	return []core_http_server.Route{
		{
			Method:  http.MethodPost,
			Path:    "/auth/login",
			Handler: h.Login,
		},
		{
			Method:  http.MethodPost,
			Path:    "/auth/logout",
			Handler: h.Logout,
		},
	}
}
