package models

import "time"

type User struct {
	Id       int64     `json:"id"`
	Name     string    `json:"username"`
	Password string    `json:"password"`
	Reg      time.Time `json:"reg"`
}

type Chat struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

type Messages struct {
	Id       int64  `json:"id"`
	ChatId   int64  `json:"chat_id"`
	SenderId int64  `json:"sender_id"`
	Message  string `json:"message"`
}
