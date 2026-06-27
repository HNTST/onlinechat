package domain

import (
	core_errors "discord/internal/core/errors"
	"fmt"
	"time"
)

type User struct {
	Id       int64     `json:"id"`
	Name     string    `json:"username"`
	Password string    `json:"password"`
	Reg      time.Time `json:"reg"`
	AvatarID *int64    `json:"avatar_id,omitempty"`
}

func NewUser(id int64, name, password string, reg time.Time) User {
	return User{
		Id:       id,
		Name:     name,
		Password: UninitializedPassword,
		Reg:      reg,
	}
}

func NewUserWithAvatar(id int64, name, password string, reg time.Time, avatarID *int64) User {
	return User{
		Id:       id,
		Name:     name,
		Password: UninitializedPassword,
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
	// когда будет регистрация сделать валидацию пароля
	return nil
}

// type UserRepository interface {
// 	GetUser(id int64) (User, error)
// 	GetUserID(username string) (int64, error)
// 	SaveUser(username, password string) (int64, error)
// }

// type UserService interface {
// 	SaveUser(username, password string) (int64, error)
// 	GetUser(user_id int64) (User, error)
// 	GetUserID(username string) (int64, error)
// }
