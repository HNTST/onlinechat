package users_transport_http

import (
	"discord/internal/core/domain"
	"time"
)

type UserDTOResponse struct {
	Id   int64     `json:"id"`
	Name string    `json:"username"`
	Reg  time.Time `json:"reg"`
}

func UserDtoFromDomain(user domain.User) UserDTOResponse {
	return UserDTOResponse{
		Id:   user.Id,
		Name: user.Name,
		Reg:  user.Reg,
	}
}
