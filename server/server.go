package server

import (
	"discord/database"
	db "discord/database"
	md "discord/models"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type serverRepo interface {
	SaveUser() http.HandlerFunc
	GetUser() http.HandlerFunc
	GetUserChats() http.HandlerFunc
	HandleConnections() http.HandlerFunc
	Broadcast(msg []byte)
}
type Server struct {
	repo      db.DBRepo
	clients   map[*websocket.Conn]bool
	clientsMu sync.Mutex
}

func New(r database.DBRepo) *Server {
	return &Server{
		repo:      r,
		clients:   make(map[*websocket.Conn]bool),
		clientsMu: sync.Mutex{},
	}
}

func (sv *Server) SaveUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			err error
		)
		var input struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		err = json.NewDecoder(r.Body).Decode(&input)
		if err != nil {
			fmt.Println("Ошибка декодирования джон запроса: ", err)
			return
		}
		id, err := sv.repo.SaveUser(input.Username, input.Password)
		if err != nil {
			fmt.Println("Ошибка сохранения в базу данных: ", err)
			return
		}
		err = json.NewEncoder(w).Encode(id)
		if err != nil {
			fmt.Println("Ошибка кодирования джон ответа: ", err)
			return
		}
	}
}
func (sv *Server) SaveChat() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			err error
		)
		var input struct {
			ChatName string `json:"chat_name"`
			Author   int64  `json:"author"`
		}
		err = json.NewDecoder(r.Body).Decode(&input)
		if err != nil {
			fmt.Println("Ошибка декодирования джон запроса: ", err)
			return
		}
		id, err := sv.repo.SaveChat(input.ChatName)
		if err != nil {
			fmt.Println("Ошибка сохранения в базу данных: ", err)
			return
		}
		err = sv.repo.SaveUserToChat(input.Author, id)
		if err != nil {
			fmt.Println("Ошибка сохранения в базу данных: ", err)
			return
		}
		err = json.NewEncoder(w).Encode(id)
		if err != nil {
			fmt.Println("Ошибка кодирования джон ответа: ", err)
			return
		}
	}
}

// --- GET
func (sv *Server) GetUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			err  error
			user md.User
		)
		var input struct {
			ID int `json:"user_id"`
		}
		err = json.NewDecoder(r.Body).Decode(&input)
		if err != nil {
			fmt.Println("Ошибка декодирования джон запроса: ", err)
			return
		}

		user, err = sv.repo.GetUser(int64(input.ID))
		if err != nil {
			fmt.Println("Ошибка сохранения в базу данных: ", err)
			return
		}
		err = json.NewEncoder(w).Encode(user)
		if err != nil {
			fmt.Println("Ошибка кодирования джон ответа: ", err)
			return
		}
	}
}
func (sv *Server) GetUserChats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			err   error
			chats []md.Chat
		)
		var input struct {
			UserID int `json:"user_id"`
		}
		err = json.NewDecoder(r.Body).Decode(&input)
		if err != nil {
			fmt.Println("Ошибка декодирования джон запроса: ", err)
			return
		}
		chats, err = sv.repo.GetUserChats(int64(input.UserID))
		if err != nil {
			fmt.Println("Ошибка сохранения в базу данных: ", err)
			return
		}
		err = json.NewEncoder(w).Encode(chats)
		if err != nil {
			fmt.Println("Ошибка кодирования джон ответа: ", err)
			return
		}
	}
}
