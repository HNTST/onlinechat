package users_transport_http

import (
	"discord/internal/core/domain"
	"time"
)

type GetUserResponse UserDTOResponse

type UserDTOResponse struct {
	Id       int64     `json:"id"`
	Name     string    `json:"username"`
	Reg      time.Time `json:"reg"`
	AvatarID *int64    `json:"avatar_id,omitempty"`
}

func UserDtoFromDomain(user domain.User) UserDTOResponse {
	return UserDTOResponse{
		Id:       user.Id,
		Name:     user.Name,
		Reg:      user.Reg,
		AvatarID: user.AvatarID,
	}
}
