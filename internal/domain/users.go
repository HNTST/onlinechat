package domain

import "time"

type User struct {
	Id       int64     `json:"id"`
	Name     string    `json:"username"`
	Password string    `json:"password"`
	Reg      time.Time `json:"reg"`
}
