package domain

import (
	"errors"
	"time"
)

var (
	ErrUserAlreadyExists = errors.New("Пользователь уже существует")
	ErrInternal          = errors.New("Внутренняя ошибка сервера")
	ErrInvalidJSON       = errors.New("Неверный формат JSON")
	ErrUserNotFound      = errors.New("Пользователь не найден")
	ErrParamMissing      = errors.New("Отсутствует обязательный параметр")
	ErrEncodingRequest   = errors.New("Ошибка кодирования JSON ответа")
)

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
	Id       int64     `json:"id"`
	ChatId   int64     `json:"chat_id"`
	SenderId int64     `json:"sender_id"`
	Message  string    `json:"message"`
	Sent_at  time.Time `json:"message"`
}

type Repository interface {
	InitDB() error

	GetUser(id int64) (User, error)
	GetUserID(username string) (int64, error)
	GetUserChats(user_id int64) ([]Chat, error)
	GetChatMessages(chat_id int64) ([]Messages, error)
	GetChatMembers(chat_id int64) ([]int64, error)

	SaveUser(username, password string) (int64, error)
	SaveChatWithAuthor(name string, user_id, target_user_id int64) (int64, error)
	SaveMessage(message Messages) error
	SaveUserToChat(user_id, chat_id int64) error
}

type Service interface {
	SaveUser(username, password string) (int64, error)
	SaveChat(chat_name string, user_id, target_user_id int64) (int64, error)
	GetUser(user_id int64) (User, error)
	GetUserChats(user_id int64) ([]Chat, error)
	GetUserMessages(chat_id int64) ([]Messages, error)
	GetUserID(username string) (int64, error)
	GetChatMembers(chat_id int64) ([]int64, error)
	SaveMessage(message Messages) error
}
