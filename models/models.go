package models

import "time"

type User struct {
	Id       int64
	Name     string
	Password string
	Reg      time.Time
}

type Chat struct {
	Id   int64
	Name string
}

type Messages struct {
	Id       int64
	ChatId   int64
	SenderId int64
	Message  string
}
