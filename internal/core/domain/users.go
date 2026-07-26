package domain

import (
	core_errors "discord/internal/core/errors"
	"fmt"
	"time"
)

type User struct {
	Id       int64     `json:"id"`
	Name     string    `json:"username"`
	Password string    `json:"-"`
	Reg      time.Time `json:"reg"`
	AvatarID *int64    `json:"avatar_id,omitempty"`
}

func NewUser(id int64, name, password string, reg time.Time) User {
	return User{
		Id:       id,
		Name:     name,
		Password: password,
		Reg:      reg,
	}
}

func NewUserWithAvatar(id int64, name, password string, reg time.Time, avatarID *int64) User {
	return User{
		Id:       id,
		Name:     name,
		Password: password,
		Reg:      reg,
		AvatarID: avatarID,
	}
}

func NewUserUninitialized(name, password string) User {
	return NewUser(int64(UninitializedId), name, password, UninitializedReg)
}

func (u *User) Validate() error {
	nameLength := len([]rune(u.Name))
	if nameLength < 3 || nameLength > 100 {
		return fmt.Errorf("не корректный `username` длинна: %d: %w", nameLength, core_errors.ErrInvalidArgument)
	}
	return nil
}
