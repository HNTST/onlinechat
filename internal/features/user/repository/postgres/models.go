package users_postgres_repository

import "time"

type UserModel struct {
	ID       int64
	Username string
	Password string
	Reg      time.Time
	AvatarID *int64
}
